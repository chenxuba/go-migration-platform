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

	if result, ok, err := svc.repo.FindInstitutionCoordinateByAddress(ctx, query); err != nil {
		return model.InstitutionGeocodeResult{}, err
	} else if ok {
		return result, nil
	}

	if result, ok := svc.resolveInstitutionCoordinateByAmap(ctx, query); ok {
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

	return model.InstitutionGeocodeResult{}, fmt.Errorf("unable to resolve coordinates")
}

func (svc *Service) resolveInstitutionCoordinateByAmap(ctx context.Context, query model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, bool) {
	key := strings.TrimSpace(svc.amapWebKey)
	if key == "" {
		return model.InstitutionGeocodeResult{}, false
	}

	params := url.Values{}
	params.Set("key", key)
	params.Set("address", query.Address)
	params.Set("city", query.City)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://restapi.amap.com/v3/geocode/geo?"+params.Encode(), nil)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}

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

	var payload struct {
		Status   string `json:"status"`
		Geocodes []struct {
			Location         string `json:"location"`
			FormattedAddress string `json:"formatted_address"`
		} `json:"geocodes"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	if payload.Status != "1" || len(payload.Geocodes) == 0 {
		return model.InstitutionGeocodeResult{}, false
	}

	parts := strings.Split(payload.Geocodes[0].Location, ",")
	if len(parts) != 2 {
		return model.InstitutionGeocodeResult{}, false
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return model.InstitutionGeocodeResult{}, false
	}

	return model.InstitutionGeocodeResult{
		Lng:             lng,
		Lat:             lat,
		Source:          "amap",
		ResolvedAddress: strings.TrimSpace(payload.Geocodes[0].FormattedAddress),
	}, true
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
