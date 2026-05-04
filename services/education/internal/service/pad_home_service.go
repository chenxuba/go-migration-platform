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
	defaultPadWeatherCity      = "深圳"
	defaultPadWeatherLatitude  = 22.5431
	defaultPadWeatherLongitude = 114.0579
	padHomeWeatherCacheTTL     = 10 * time.Minute
)

var sharedPadHomeWeatherCache = struct {
	sync.Mutex
	weather    model.PadHomeWeather
	expiresAt  time.Time
	refreshing bool
}{}

func (svc *Service) GetPadHomeSummary(userID int64) (model.PadHomeSummaryVO, error) {
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
	stats, err := svc.repo.GetPadHomeAssessmentStats(ctx, instID, "", now)
	if err != nil {
		return model.PadHomeSummaryVO{}, err
	}

	schedules, err := svc.ListTeachingSchedules(userID, model.TeachingScheduleListQueryDTO{
		StartDate:     today,
		EndDate:       today,
		SortDirection: "asc",
	})
	if err != nil {
		return model.PadHomeSummaryVO{}, err
	}

	return model.PadHomeSummaryVO{
		Date:            today,
		Weekday:         chineseWeekday(now.Weekday()),
		AssessmentStats: stats,
		Schedule:        padHomeScheduleItems(schedules, now),
		Weather:         svc.padHomeWeather(ctx),
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

func (svc *Service) padHomeWeather(_ context.Context) model.PadHomeWeather {
	now := time.Now()
	sharedPadHomeWeatherCache.Lock()
	cached := sharedPadHomeWeatherCache.weather
	if !sharedPadHomeWeatherCache.expiresAt.IsZero() && now.Before(sharedPadHomeWeatherCache.expiresAt) {
		sharedPadHomeWeatherCache.Unlock()
		return cached
	}
	fallback := cached
	if strings.TrimSpace(fallback.City) == "" {
		fallback = fallbackPadHomeWeather()
	}
	if !sharedPadHomeWeatherCache.refreshing {
		sharedPadHomeWeatherCache.refreshing = true
		go refreshPadHomeWeatherCache()
	}
	sharedPadHomeWeatherCache.Unlock()
	return fallback
}

func fallbackPadHomeWeather() model.PadHomeWeather {
	return model.PadHomeWeather{
		City:        padWeatherCity(),
		Condition:   "sunny",
		DisplayName: "晴",
		Source:      "fallback",
	}
}

func refreshPadHomeWeatherCache() {
	weather, err := fetchPadHomeWeather(context.Background())
	sharedPadHomeWeatherCache.Lock()
	defer sharedPadHomeWeatherCache.Unlock()
	sharedPadHomeWeatherCache.refreshing = false
	if err != nil {
		return
	}
	sharedPadHomeWeatherCache.weather = weather
	sharedPadHomeWeatherCache.expiresAt = time.Now().Add(padHomeWeatherCacheTTL)
}

func fetchPadHomeWeather(ctx context.Context) (model.PadHomeWeather, error) {
	location, err := resolvePadWeatherLocation(ctx)
	if err != nil {
		return model.PadHomeWeather{}, err
	}
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

func resolvePadWeatherLocation(ctx context.Context) (padWeatherLocation, error) {
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
	query := endpoint.Query()
	query.Set("name", city)
	query.Set("count", "1")
	query.Set("language", "zh")
	if country := strings.TrimSpace(os.Getenv("PAD_WEATHER_COUNTRY_CODE")); country != "" {
		query.Set("countryCode", country)
	} else {
		query.Set("countryCode", "CN")
	}
	endpoint.RawQuery = query.Encode()

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

func padWeatherCoordinatesFromEnv() (float64, float64, bool) {
	latitude, latErr := strconv.ParseFloat(strings.TrimSpace(os.Getenv("PAD_WEATHER_LATITUDE")), 64)
	longitude, lngErr := strconv.ParseFloat(strings.TrimSpace(os.Getenv("PAD_WEATHER_LONGITUDE")), 64)
	if latErr != nil || lngErr != nil || latitude == 0 || longitude == 0 {
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
