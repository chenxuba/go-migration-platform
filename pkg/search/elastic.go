package search

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ElasticClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewElasticClient(baseURL, username, password string) *ElasticClient {
	return &ElasticClient{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (client *ElasticClient) Health() (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, client.baseURL+"/_cluster/health", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(client.username, client.password)

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("es health status %d: %s", resp.StatusCode, string(body))
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (client *ElasticClient) IndexExists(index string) (bool, error) {
	req, err := http.NewRequest(http.MethodHead, client.baseURL+"/"+index, nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(client.username, client.password)

	resp, err := client.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return false, fmt.Errorf("index exists status %d", resp.StatusCode)
	}
	return true, nil
}

func (client *ElasticClient) EnsureIntentStudentIndex(index string) error {
	exists, err := client.IndexExists(index)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":               map[string]any{"type": "keyword"},
				"instId":           map[string]any{"type": "keyword"},
				"stuName":          map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
				"mobile":           map[string]any{"type": "keyword"},
				"studentStatus":    map[string]any{"type": "integer"},
				"intentLevel":      map[string]any{"type": "integer"},
				"followUpStatus":   map[string]any{"type": "integer"},
				"channelId":        map[string]any{"type": "keyword"},
				"createTime":       map[string]any{"type": "date"},
				"followUpTime":     map[string]any{"type": "date"},
				"nextFollowUpTime": map[string]any{"type": "date"},
			},
		},
	}

	payload, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, client.baseURL+"/"+index, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("create index status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (client *ElasticClient) DeleteIndex(index string) error {
	req, err := http.NewRequest(http.MethodDelete, client.baseURL+"/"+index, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(client.username, client.password)

	resp, err := client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("delete index status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (client *ElasticClient) BulkIndex(index string, docs []map[string]any) error {
	if len(docs) == 0 {
		return nil
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	for _, doc := range docs {
		meta := map[string]any{
			"index": map[string]any{
				"_index": index,
				"_id":    doc["id"],
			},
		}
		if err := encoder.Encode(meta); err != nil {
			return err
		}
		if err := encoder.Encode(doc); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodPost, client.baseURL+"/_bulk?refresh=true", &buffer)
	if err != nil {
		return err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("bulk index status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Errors bool `json:"errors"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.Errors {
		return fmt.Errorf("bulk index returned errors")
	}
	return nil
}

func (client *ElasticClient) DeleteIntentStudentsByInstID(index string, instID int64) (bool, error) {
	exists, err := client.IndexExists(index)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	payload, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"instId": fmt.Sprintf("%d", instID),
			},
		},
	})
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, client.baseURL+"/"+index+"/_delete_by_query?refresh=true", bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return false, fmt.Errorf("delete by query status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Deleted  int   `json:"deleted"`
		Failures []any `json:"failures"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	if len(result.Failures) > 0 {
		return false, fmt.Errorf("delete by query returned failures")
	}
	return result.Deleted > 0, nil
}
