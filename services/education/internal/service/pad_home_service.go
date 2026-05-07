package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	defaultPadWeatherCity      = "北京"
	defaultPadWeatherLatitude  = 39.9042
	defaultPadWeatherLongitude = 116.4074
	padHomeWeatherCacheTTL     = 10 * time.Minute
)

type padHomeWeatherCacheEntry struct {
	weather    model.PadHomeWeather
	expiresAt  time.Time
	refreshing bool
}

var sharedPadHomeWeatherCache = struct {
	sync.Mutex
	entries map[string]padHomeWeatherCacheEntry
}{
	entries: map[string]padHomeWeatherCacheEntry{},
}

func (svc *Service) GetPadHomeSummary(userID int64, query model.PadHomeSummaryQueryDTO) (model.PadHomeSummaryVO, error) {
	if svc.repo == nil {
		return model.PadHomeSummaryVO{}, errors.New("education repository is not configured")
	}
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PadHomeSummaryVO{}, errors.New("no institution context")
		}
		return model.PadHomeSummaryVO{}, err
	}

	now := time.Now().In(padHomeLocation())
	today := now.Format("2006-01-02")
	var (
		stats     model.PadHomeAssessmentStats
		schedules []model.TeachingScheduleVO
		weather   model.PadHomeWeather
		statsErr  error
		schedErr  error
	)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		stats, statsErr = svc.repo.GetPadHomeAssessmentStats(ctx, instID, "", now)
	}()
	go func() {
		defer wg.Done()
		schedules, schedErr = svc.repo.ListPadHomeSchedules(ctx, instID, today)
		if schedErr != nil {
			return
		}
		if err := svc.repo.FillTeachingScheduleCallStatus(ctx, instID, schedules); err != nil {
			schedErr = err
		}
	}()
	go func() {
		defer wg.Done()
		weather = svc.padHomeWeather(ctx, instID, query)
	}()
	wg.Wait()
	if statsErr != nil {
		return model.PadHomeSummaryVO{}, statsErr
	}
	if schedErr != nil {
		return model.PadHomeSummaryVO{}, schedErr
	}

	return model.PadHomeSummaryVO{
		Date:            today,
		Weekday:         chineseWeekday(now.Weekday()),
		AssessmentStats: stats,
		Schedule:        padHomeScheduleItems(schedules, now),
		Weather:         weather,
	}, nil
}

func padHomeLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return loc
	}
	return time.Local
}

func chineseWeekday(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return "星期一"
	case time.Tuesday:
		return "星期二"
	case time.Wednesday:
		return "星期三"
	case time.Thursday:
		return "星期四"
	case time.Friday:
		return "星期五"
	case time.Saturday:
		return "星期六"
	default:
		return "星期日"
	}
}

func padHomeScheduleItems(schedules []model.TeachingScheduleVO, now time.Time) []model.PadHomeScheduleItem {
	limit := len(schedules)
	if limit > 4 {
		limit = 4
	}
	items := make([]model.PadHomeScheduleItem, 0, limit)
	for i := 0; i < limit; i++ {
		schedule := schedules[i]
		title := strings.TrimSpace(schedule.LessonName)
		if className := strings.TrimSpace(schedule.TeachingClassName); className != "" {
			if title != "" {
				title += " · " + className
			} else {
				title = className
			}
		}
		if title == "" {
			title = "未命名课程"
		}
		place := strings.TrimSpace(schedule.ClassroomName)
		if place == "" {
			place = "未分配教室"
		}
		items = append(items, model.PadHomeScheduleItem{
			Time:  schedule.StartAt.In(now.Location()).Format("15:04"),
			Title: title,
			Place: place,
			State: padHomeScheduleState(schedule, now),
		})
	}
	return items
}

func padHomeScheduleState(schedule model.TeachingScheduleVO, now time.Time) string {
	startAt := schedule.StartAt.In(now.Location())
	endAt := schedule.EndAt.In(now.Location())
	if schedule.CallStatus == 2 || schedule.CallStatus == 3 {
		return "已点名"
	}
	if !now.Before(endAt) {
		return "已结束"
	}
	if !now.Before(startAt) && now.Before(endAt) {
		return "进行中"
	}
	if startAt.Sub(now) <= 30*time.Minute {
		return "即将开始"
	}
	return "未开始"
}

func (svc *Service) padHomeWeather(ctx context.Context, instID int64, query model.PadHomeSummaryQueryDTO) model.PadHomeWeather {
	location, err := svc.resolvePadWeatherLocation(ctx, instID, query)
	if err != nil {
		return fallbackPadHomeWeather(query.City)
	}
	cacheKey := location.cacheKey()
	now := time.Now()
	sharedPadHomeWeatherCache.Lock()
	entry := sharedPadHomeWeatherCache.entries[cacheKey]
	if !entry.expiresAt.IsZero() && now.Before(entry.expiresAt) {
		sharedPadHomeWeatherCache.Unlock()
		return entry.weather
	}
	stale := entry.weather
	if strings.TrimSpace(stale.City) != "" {
		if !entry.refreshing {
			entry.refreshing = true
			sharedPadHomeWeatherCache.entries[cacheKey] = entry
			go refreshPadHomeWeatherCache(cacheKey, location)
		}
		sharedPadHomeWeatherCache.Unlock()
		return stale
	}
	sharedPadHomeWeatherCache.Unlock()

	weather, err := fetchPadHomeWeather(ctx, location)
	if err != nil {
		return fallbackPadHomeWeather(location.city)
	}
	storePadHomeWeatherCache(cacheKey, weather)
	return weather
}

func fallbackPadHomeWeather(city string) model.PadHomeWeather {
	if strings.TrimSpace(city) == "" {
		city = "当前位置"
	}
	return model.PadHomeWeather{
		City:        city,
		Condition:   "sunny",
		DisplayName: "晴",
		Source:      "fallback",
	}
}

func refreshPadHomeWeatherCache(cacheKey string, location padWeatherLocation) {
	weather, err := fetchPadHomeWeather(context.Background(), location)
	sharedPadHomeWeatherCache.Lock()
	defer sharedPadHomeWeatherCache.Unlock()
	entry := sharedPadHomeWeatherCache.entries[cacheKey]
	entry.refreshing = false
	if err != nil {
		sharedPadHomeWeatherCache.entries[cacheKey] = entry
		return
	}
	entry.weather = weather
	entry.expiresAt = time.Now().Add(padHomeWeatherCacheTTL)
	sharedPadHomeWeatherCache.entries[cacheKey] = entry
}

func storePadHomeWeatherCache(cacheKey string, weather model.PadHomeWeather) {
	sharedPadHomeWeatherCache.Lock()
	defer sharedPadHomeWeatherCache.Unlock()
	sharedPadHomeWeatherCache.entries[cacheKey] = padHomeWeatherCacheEntry{
		weather:   weather,
		expiresAt: time.Now().Add(padHomeWeatherCacheTTL),
	}
}

func fetchPadHomeWeather(ctx context.Context, location padWeatherLocation) (model.PadHomeWeather, error) {
	endpoint, err := url.Parse("https://api.open-meteo.com/v1/forecast")
	if err != nil {
		return model.PadHomeWeather{}, err
	}
	query := endpoint.Query()
	query.Set("latitude", strconv.FormatFloat(location.latitude, 'f', 4, 64))
	query.Set("longitude", strconv.FormatFloat(location.longitude, 'f', 4, 64))
	query.Set("current", "temperature_2m,weather_code,cloud_cover,precipitation,rain,showers")
	query.Set("timezone", "auto")
	endpoint.RawQuery = query.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return model.PadHomeWeather{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.PadHomeWeather{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return model.PadHomeWeather{}, fmt.Errorf("weather api status %d", resp.StatusCode)
	}

	var payload openMeteoForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return model.PadHomeWeather{}, err
	}
	condition, displayName := normalizePadWeather(payload.Current.WeatherCode, payload.Current.Precipitation, payload.Current.Rain, payload.Current.Showers)
	return model.PadHomeWeather{
		City:        location.city,
		Condition:   condition,
		DisplayName: displayName,
		Temperature: math.Round(payload.Current.Temperature*10) / 10,
		UpdatedAt:   strings.TrimSpace(payload.Current.Time),
		Source:      "open-meteo",
	}, nil
}

type padWeatherLocation struct {
	city      string
	latitude  float64
	longitude float64
}

func (location padWeatherLocation) cacheKey() string {
	return fmt.Sprintf("%s|%.4f,%.4f", strings.TrimSpace(location.city), location.latitude, location.longitude)
}

func (svc *Service) resolvePadWeatherLocation(ctx context.Context, instID int64, query model.PadHomeSummaryQueryDTO) (padWeatherLocation, error) {
	if query.HasCoordinates && validPadWeatherCoordinates(query.Latitude, query.Longitude) {
		city := strings.TrimSpace(query.City)
		if city == "" {
			city = "当前位置"
		}
		return padWeatherLocation{
			city:      city,
			latitude:  query.Latitude,
			longitude: query.Longitude,
		}, nil
	}

	if svc.repo != nil && instID > 0 {
		location, err := svc.repo.GetInstitutionWeatherLocation(ctx, instID)
		if err == nil && validPadWeatherCoordinates(location.Latitude, location.Longitude) {
			city := strings.TrimSpace(location.City)
			if city == "" {
				city = "机构地址"
			}
			return padWeatherLocation{
				city:      city,
				latitude:  location.Latitude,
				longitude: location.Longitude,
			}, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return padWeatherLocation{}, err
		}
	}

	if latitude, longitude, ok := padWeatherCoordinatesFromEnv(); ok {
		return padWeatherLocation{
			city:      padWeatherCity(),
			latitude:  latitude,
			longitude: longitude,
		}, nil
	}

	city := padWeatherCity()
	endpoint, err := url.Parse("https://geocoding-api.open-meteo.com/v1/search")
	if err != nil {
		return padWeatherLocation{}, err
	}
	values := endpoint.Query()
	values.Set("name", city)
	values.Set("count", "1")
	values.Set("language", "zh")
	if country := strings.TrimSpace(os.Getenv("PAD_WEATHER_COUNTRY_CODE")); country != "" {
		values.Set("countryCode", country)
	} else {
		values.Set("countryCode", "CN")
	}
	endpoint.RawQuery = values.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return padWeatherLocation{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fallbackPadWeatherLocation(city), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fallbackPadWeatherLocation(city), nil
	}
	var payload openMeteoGeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fallbackPadWeatherLocation(city), nil
	}
	if len(payload.Results) == 0 {
		return fallbackPadWeatherLocation(city), nil
	}
	first := payload.Results[0]
	if first.Latitude == 0 && first.Longitude == 0 {
		return fallbackPadWeatherLocation(city), nil
	}
	if strings.TrimSpace(first.Name) != "" {
		city = first.Name
	}
	return padWeatherLocation{
		city:      city,
		latitude:  first.Latitude,
		longitude: first.Longitude,
	}, nil
}

func validPadWeatherCoordinates(latitude, longitude float64) bool {
	return latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180 && !(latitude == 0 && longitude == 0)
}

func padWeatherCoordinatesFromEnv() (float64, float64, bool) {
	latitude, latErr := strconv.ParseFloat(strings.TrimSpace(os.Getenv("PAD_WEATHER_LATITUDE")), 64)
	longitude, lngErr := strconv.ParseFloat(strings.TrimSpace(os.Getenv("PAD_WEATHER_LONGITUDE")), 64)
	if latErr != nil || lngErr != nil || !validPadWeatherCoordinates(latitude, longitude) {
		return 0, 0, false
	}
	return latitude, longitude, true
}

func padWeatherCity() string {
	if city := strings.TrimSpace(os.Getenv("PAD_WEATHER_CITY")); city != "" {
		return city
	}
	return defaultPadWeatherCity
}

func fallbackPadWeatherLocation(city string) padWeatherLocation {
	if strings.TrimSpace(city) == "" {
		city = defaultPadWeatherCity
	}
	return padWeatherLocation{
		city:      city,
		latitude:  defaultPadWeatherLatitude,
		longitude: defaultPadWeatherLongitude,
	}
}

func normalizePadWeather(weatherCode int, precipitation, rain, showers float64) (string, string) {
	if precipitation > 0 || rain > 0 || showers > 0 {
		return "rain", "下雨"
	}
	switch weatherCode {
	case 0:
		return "sunny", "晴"
	case 1, 2:
		return "partly_cloudy", "多云"
	case 3, 45, 48:
		return "overcast", "阴"
	case 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82, 95, 96, 99:
		return "rain", "下雨"
	default:
		return "cloudy", "多云"
	}
}

type openMeteoForecastResponse struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature   float64 `json:"temperature_2m"`
		WeatherCode   int     `json:"weather_code"`
		Precipitation float64 `json:"precipitation"`
		Rain          float64 `json:"rain"`
		Showers       float64 `json:"showers"`
	} `json:"current"`
}

type openMeteoGeocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}
