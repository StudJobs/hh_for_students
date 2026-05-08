package esclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	IndexProfiles  = "profiles"
	IndexVacancies = "vacancies"
)

type Client struct {
	es *elasticsearch.Client
}

func New(url string) (*Client, error) {
	cfg := elasticsearch.Config{Addresses: []string{url}}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: new client: %w", err)
	}
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: info: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch: info: %s", res.String())
	}
	return &Client{es: es}, nil
}

// EnsureIndex создаёт индекс с заданным mapping, если его нет.
// recreate=true — удаляет существующий и пересоздаёт.
func (c *Client) EnsureIndex(ctx context.Context, name, mapping string, recreate bool) error {
	exists, err := c.indexExists(ctx, name)
	if err != nil {
		return err
	}
	if exists && recreate {
		if err := c.deleteIndex(ctx, name); err != nil {
			return err
		}
		exists = false
	}
	if exists {
		return nil
	}
	res, err := c.es.Indices.Create(name, c.es.Indices.Create.WithContext(ctx), c.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("elasticsearch: create index %s: %w", name, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch: create index %s: %s", name, string(body))
	}
	return nil
}

func (c *Client) indexExists(ctx context.Context, name string) (bool, error) {
	res, err := c.es.Indices.Exists([]string{name}, c.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("elasticsearch: exists index %s: %w", name, err)
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("elasticsearch: exists index %s: status %d", name, res.StatusCode)
	}
}

func (c *Client) deleteIndex(ctx context.Context, name string) error {
	res, err := c.es.Indices.Delete([]string{name}, c.es.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elasticsearch: delete index %s: %w", name, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch: delete index %s: %s", name, string(body))
	}
	return nil
}

// Index сохраняет документ. Идемпотентно: на повторный вызов с тем же id перезаписывает.
func (c *Client) Index(ctx context.Context, index, id string, doc []byte) error {
	res, err := c.es.Index(
		index,
		bytes.NewReader(doc),
		c.es.Index.WithDocumentID(id),
		c.es.Index.WithContext(ctx),
		c.es.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch: index %s/%s: %w", index, id, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch: index %s/%s: %s", index, id, string(body))
	}
	return nil
}

// Delete удаляет документ из индекса. Если документа нет — не ошибка.
func (c *Client) Delete(ctx context.Context, index, id string) error {
	res, err := c.es.Delete(index, id, c.es.Delete.WithContext(ctx), c.es.Delete.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("elasticsearch: delete %s/%s: %w", index, id, err)
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch: delete %s/%s: %s", index, id, string(body))
	}
	return nil
}

// Search выполняет произвольный JSON-запрос к индексу и возвращает сырое тело ответа.
func (c *Client) Search(ctx context.Context, index string, body []byte) ([]byte, error) {
	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(index),
		c.es.Search.WithBody(bytes.NewReader(body)),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: search %s: %w", index, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch: search %s read body: %w", index, err)
	}
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch: search %s: %s", index, string(raw))
	}
	return raw, nil
}
