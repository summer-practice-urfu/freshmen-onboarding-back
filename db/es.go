package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
)

type EsDb struct {
	client *elasticsearch.Client
	logger *log.Logger
}

func NewEsDb(logger *log.Logger) *EsDb {
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}
	res, err := client.Info()
	if err != nil {
		logger.Fatal("Can't connect to es")
	}
	defer res.Body.Close()
	logger.Println("Connected to es")
	return &EsDb{client: client, logger: logger}
}

type ESSearchResponse struct {
	Hits ESSearchResponseHits `json:"hits"`
}

type ESSearchResponseHits struct {
	Total ESSearchResponseTotal       `json:"total"`
	Hits  []ESSearchResponseHitsInner `json:"hits"`
}

type ESSearchResponseTotal struct {
	Value int `json:"value"`
}

type ESSearchResponseHitsInner struct {
	Index  *string         `json:"_index"`
	Id     *string         `json:"_id"`
	Score  *float32        `json:"_score"`
	Source json.RawMessage `json:"_source"`
}

func (e *EsDb) Search(index, query string, fields []string, size, page int) (*ESSearchResponse, error) {
	from := size * (page - 1)
	if page > 1 {
		from += 1
	}

	body := map[string]interface{}{
		"size": size,
		"from": from,
	}
	e.logger.Println("Request body to es: ", body)
	if query != "" {
		body["query"] = map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": fields,
			},
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(bodyBytes),
	}
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("response error")
	}

	var response *ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (e *EsDb) Index(index string, id string, doc any) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return errors.New("error marshaling doc")
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	return e.withoutResponse(req)
}

func (e *EsDb) Update(index, id string, doc any) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return errors.New("error marshaling doc")
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	return e.withoutResponse(req)
}

func (e *EsDb) Delete(index, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	return e.withoutResponse(req)
}

func (e *EsDb) withoutResponse(req esapi.Request) error {
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		return errors.New("error in request")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New("error in response")
	}

	return nil
}
