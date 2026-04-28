package search

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
)

const taskPollInterval = 100 * time.Millisecond

type Client struct {
	host   string
	apiKey string
	client meilisearch.ServiceManager
}

func NewClient(host, apiKey string) *Client {
	host = strings.TrimRight(strings.TrimSpace(host), "/")
	if host == "" {
		host = "http://127.0.0.1:7700"
	}
	options := make([]meilisearch.Option, 0, 1)
	if strings.TrimSpace(apiKey) != "" {
		options = append(options, meilisearch.WithAPIKey(apiKey))
	}
	return &Client{
		host:   host,
		apiKey: apiKey,
		client: meilisearch.New(host, options...),
	}
}

func (client *Client) Health() (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	health, err := client.client.HealthWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"host":   client.host,
		"status": health.Status,
	}, nil
}

func (client *Client) IndexExists(index string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.client.GetIndexWithContext(ctx, index)
	if err == nil {
		return true, nil
	}
	if isNotFound(err) {
		return false, nil
	}
	return false, err
}

func (client *Client) EnsureIntentStudentIndex(index string) error {
	exists, err := client.IndexExists(index)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !exists {
		task, err := client.client.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
			Uid:        index,
			PrimaryKey: "id",
		})
		if err != nil {
			return err
		}
		if err := client.waitForTask(ctx, task.TaskUID); err != nil {
			return err
		}
	}
	task, err := client.client.Index(index).UpdateSettingsWithContext(ctx, &meilisearch.Settings{
		SearchableAttributes: []string{"stuName", "mobile"},
		FilterableAttributes: []string{"instId", "studentStatus", "intentLevel", "followUpStatus", "channelId"},
		SortableAttributes:   []string{"createTime", "followUpTime", "nextFollowUpTime"},
	})
	if err != nil {
		return err
	}
	return client.waitForTask(ctx, task.TaskUID)
}

func (client *Client) DeleteIndex(index string) error {
	exists, err := client.IndexExists(index)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	task, err := client.client.DeleteIndexWithContext(ctx, index)
	if err != nil {
		return err
	}
	return client.waitForTask(ctx, task.TaskUID)
}

func (client *Client) BulkIndex(index string, docs []map[string]any) error {
	if len(docs) == 0 {
		return nil
	}
	if err := client.EnsureIntentStudentIndex(index); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	primaryKey := "id"
	task, err := client.client.Index(index).AddDocumentsWithContext(ctx, docs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}
	return client.waitForTask(ctx, task.TaskUID)
}

func (client *Client) DeleteIntentStudentsByInstID(index string, instID int64) (bool, error) {
	exists, err := client.IndexExists(index)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	task, err := client.client.Index(index).DeleteDocumentsByFilterWithContext(ctx, fmt.Sprintf("instId = \"%d\"", instID), nil)
	if err != nil {
		return false, err
	}
	if err := client.waitForTask(ctx, task.TaskUID); err != nil {
		return false, err
	}
	return true, nil
}

func (client *Client) waitForTask(ctx context.Context, taskUID int64) error {
	task, err := client.client.WaitForTaskWithContext(ctx, taskUID, taskPollInterval)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("meilisearch task returned empty result")
	}
	if task.Status == meilisearch.TaskStatusFailed || task.Status == meilisearch.TaskStatusCanceled {
		if task.Error.Message != "" {
			return fmt.Errorf("meilisearch task %d %s: %s", taskUID, task.Status, task.Error.Message)
		}
		return fmt.Errorf("meilisearch task %d %s", taskUID, task.Status)
	}
	return nil
}

func isNotFound(err error) bool {
	var meiliErr *meilisearch.Error
	return errors.As(err, &meiliErr) && meiliErr.StatusCode == http.StatusNotFound
}
