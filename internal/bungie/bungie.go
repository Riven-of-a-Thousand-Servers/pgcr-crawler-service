package bungie

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	types "github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/types"
)

type PgcrClient interface {
	FetchPgcr(instanceId int64) (*types.PostGameCarnageReportResponse, error)
}

type BungieClient struct {
	Client *http.Client
	ApiKey string
	Host   string
}

var (
	statsURI     = "http://%s/Platform/Destiny2/Stats/PostGameCarnageReport/%d/"
	apiKeyHeader = "x-api-key"
)

func NewBungieClient(apiKey, host string) (*BungieClient, error) {
	if apiKey == "" {
		return nil, errors.New("Api key is empty")
	}

	if host == "" {
		return nil, errors.New("Host is empty")
	}

	client := http.Client{}
	return &BungieClient{
		Client: &client,
		ApiKey: apiKey,
		Host:   host,
	}, nil
}

func (p *BungieClient) FetchPgcr(instanceId int64) (*types.PostGameCarnageReportResponse, error) {
	uri := fmt.Sprintf(statsURI, p.Host, instanceId)
	log.Printf("URI: %s", uri)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request for pgcr [%d]: %v", instanceId, err)
	}
	log.Printf("BungieClient: %#v", p)
	request.Header.Set(apiKeyHeader, p.ApiKey)

	response, err := p.Client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode > 299 {
		return nil, fmt.Errorf("Response for PGCR [%d] returned an error response: [%s]", instanceId, response.Status)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body for PGCR [%d]: %v", instanceId, err)
	}

	var pgcr types.PostGameCarnageReportResponse
	if err := json.Unmarshal(body, &pgcr); err != nil {
		return nil, fmt.Errorf("Error unmarshalling JSON response for PGCR [%d]: %v", instanceId, err)
	}

	log.Printf("Successfuly response from Bungie for pgcr [%d]: %#v", instanceId, pgcr)
	return &pgcr, nil
}
