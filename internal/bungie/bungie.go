package bungie

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	types "github.com/Riven-of-a-Thousand-Servers/rivenbot-commons/pkg/types"
)

type BungieClient interface {
	FetchPgcr(instanceId int64) (*types.PostGameCarnageReportResponse, error)
}

type PgcrClient struct {
	Client *http.Client
	ApiKey string
	Host   string
}

var (
	statsURI     = "https://%s/Platform/Destiny2/Stats/PostGameCarnageReport/%s/"
	apiKeyHeader = "x-api-key"
)

func (p *PgcrClient) FetchPgcr(instanceId int64) (*types.PostGameCarnageReportResponse, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf(statsURI, p.Host, instanceId), nil)
	request.Header.Add(apiKeyHeader, p.ApiKey)

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
	return &pgcr, nil
}
