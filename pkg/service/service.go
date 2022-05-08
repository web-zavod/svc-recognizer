package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"

	"github.com/web-zavod/svc-recognizer/pkg/service/models"
	"github.com/web-zavod/svc-recognizer/pkg/service/queries"
)

type Service interface {
	IndexCategory(models.Category) error
	SearchCategory(string) (string, error)
	CreateIndex() error
	DeleteIndex() error
}

type service struct {
	es    *elasticsearch.Client
	index string
}

type dict map[string]interface{}

func NewService(es *elasticsearch.Client, index string) Service {
	return &service{es, index}
}

func (s *service) SearchCategory(text string) (string, error) {
	esreq := esapi.SearchRequest{
		Index: []string{s.index},
		Body:  queries.SearchCategories(text),
	}
	res, err := esreq.Do(context.Background(), s.es)
	if err != nil {
		return "", err
	}

	var m models.SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("Error parsing the response body: %s", err)
	}

	if m.Hits.Total.Value == 0 {
		return "", fmt.Errorf("Category not found for pattern: %s", text)
	}

	return m.Hits.Hits[0].Source.Query.Name, nil
}

// Indexes category in ElasticSearch
func (s *service) IndexCategory(category models.Category) error {
	body := map[string]interface{}{
		"id": category.ID,
		"query": map[string]interface{}{
			"name": category.Category,
		},
	}
	esreq := esapi.IndexRequest{
		Index:      s.index,
		Body:       esutil.NewJSONReader(body),
		DocumentID: category.ID,
		Refresh:    "true",
	}

	res, err := esreq.Do(context.Background(), s.es)
	return s.handleResponseWithoutProcessing(res, err)
}

// Creates a category index in ElasticSearch
func (s *service) CreateIndex() error {
	esreq := esapi.IndicesCreateRequest{Index: s.index, Body: queries.CreateCategoriesIndex()}
	res, err := esreq.Do(context.Background(), s.es)
	if err != nil {
		return err
	}

	return s.handleResponseWithoutProcessing(res, err)
}

// Deletes a category index created in ElasticSearch.
func (s *service) DeleteIndex() error {
	esreq := esapi.IndicesDeleteRequest{
		Index: []string{s.index},
	}
	res, err := esreq.Do(context.Background(), s.es)
	return s.handleResponseWithoutProcessing(res, err)
}

// Handles the response from ElasticSearch where
// processing of a successful result is not required.
func (s *service) handleResponseWithoutProcessing(res *esapi.Response, err error) error {
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return s.handleErrorResponse(res)
	}
	return nil
}

// Handles the standard error response from ElasticSearch.
func (s *service) handleErrorResponse(res *esapi.Response) error {
	var error models.ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&error); err != nil {
		return fmt.Errorf("Error parsing the response body: %s", err)
	}
	return fmt.Errorf("[%s] %s: %s",
		res.Status(),
		error.Info.Type,
		error.Info.Reason,
	)
}
