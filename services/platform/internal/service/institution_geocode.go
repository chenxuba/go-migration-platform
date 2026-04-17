package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/platform/internal/model"
)

func (svc *Service) resolveInstitutionCoordinate(ctx context.Context, input model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, error) {
	query := model.InstitutionGeocodeQuery{
		Province: strings.TrimSpace(input.Province),
		City:     strings.TrimSpace(input.City),
		Region:   strings.TrimSpace(input.Region),
		Address:  strings.TrimSpace(input.Address),
	}
	var amapErr error

	if result, ok, err := svc.repo.FindInstitutionCoordinateByAddress(ctx, query); err != nil {
		return model.InstitutionGeocodeResult{}, err
	} else if ok {
		return result, nil
	}

	if result, ok, err := svc.resolveInstitutionCoordinateByAmap(ctx, query); err != nil {
		amapErr = err
	} else if ok {
		return result, nil
	}

	if result, ok := svc.resolveInstitutionCoordinateByNominatim(ctx, query); ok {
		return result, nil
	}

	if result, ok, err := svc.repo.FindInstitutionCoordinateFallback(ctx, query); err != nil {
		return model.InstitutionGeocodeResult{}, err
	} else if ok {
		return result, nil
	}

	if amapErr != nil {
		return model.InstitutionGeocodeResult{}, amapErr
	}

	return model.InstitutionGeocodeResult{}, fmt.Errorf("地址解析失败，请检查省市区和详细地址是否准确")
}

func (svc *Service) resolveInstitutionCoordinateByAmap(ctx context.Context, query model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, bool, error) {
	key := strings.TrimSpace(svc.amapWebKey)
	if key == "" {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	params := url.Values{}
	params.Set("key", key)
	params.Set("address", query.Address)
	params.Set("city", query.City)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://restapi.amap.com/v3/geocode/geo?"+params.Encode(), nil)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	var payload struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Infocode string `json:"infocode"`
		Geocodes []struct {
			Location         string `json:"location"`
			FormattedAddress string `json:"formatted_address"`
		} `json:"geocodes"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}
	if payload.Status != "1" {
		if payload.Infocode == "10009" || strings.EqualFold(strings.TrimSpace(payload.Info), "USERKEY_PLAT_NOMATCH") {
			return model.InstitutionGeocodeResult{}, false, fmt.Errorf("当前高德 Key 与 Web 服务平台类型不匹配，请更换为 Web 服务 Key")
		}
		return model.InstitutionGeocodeResult{}, false, nil
	}
	if len(payload.Geocodes) == 0 {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	parts := strings.Split(payload.Geocodes[0].Location, ",")
	if len(parts) != 2 {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false, nil
	}

	return model.InstitutionGeocodeResult{
		Lng:             lng,
		Lat:             lat,
		Source:          "amap",
		ResolvedAddress: strings.TrimSpace(payload.Geocodes[0].FormattedAddress),
	}, true, nil
}

func (svc *Service) resolveInstitutionCoordinateByNominatim(ctx context.Context, query model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, bool) {
	params := url.Values{}
	params.Set("format", "jsonv2")
	params.Set("limit", "1")
	params.Set("q", buildInstitutionAddress(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://nominatim.openstreetmap.org/search?"+params.Encode(), nil)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	req.Header.Set("User-Agent", "go-migration-platform/1.0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}

	var payload []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	if len(payload) == 0 {
		return model.InstitutionGeocodeResult{}, false
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(payload[0].Lon), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(payload[0].Lat), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}

	return model.InstitutionGeocodeResult{
		Lng:             lng,
		Lat:             lat,
		Source:          "nominatim",
		ResolvedAddress: strings.TrimSpace(payload[0].DisplayName),
	}, true
}

func buildInstitutionAddress(query model.InstitutionGeocodeQuery) string {
	return strings.TrimSpace(query.Province + query.City + query.Region + query.Address)
}
