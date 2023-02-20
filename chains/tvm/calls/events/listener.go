package events

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

const (
	EventsEndpoint   = "events"
	DepositEventName = "Deposit"
)

type Fetcher struct {
	baseUrl string
	apiKey  string
}

func NewFetcher(baseUrl, apiKey string) *Fetcher {
	return &Fetcher{baseUrl: baseUrl, apiKey: apiKey}
}

func (l *Fetcher) FetchDeposits(ctx context.Context, contractAddress address.Address, startTime, endTime *time.Time) ([]Deposit, error) {
	// https://api.shasta.trongrid.io/v1/contracts/TZ2xAEKqHup6hzEQGPhmXQFiXBDBQxSVZG/events?max_block_timestamp=1676642178000&min_block_timestamp=1676642178000&only_confirmed=true&event_name=Deposit
	u := fmt.Sprintf(
		"%s/contracts/%s/%s",
		l.baseUrl,
		contractAddress.String(),
		EventsEndpoint,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	query := url.Values{
		"event_name":     []string{DepositEventName},
		"only_confirmed": []string{"true"},
	}
	if startTime != nil {
		query.Set("min_block_timestamp", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if endTime != nil {
		query.Set("max_block_timestamp", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("TRON-PRO-API-KEY", l.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code doesn't equal 200")
	}
	var responseStruct response
	if err := json.NewDecoder(resp.Body).Decode(&responseStruct); err != nil {
		return nil, err
	}
	if !responseStruct.Success {
		return nil, fmt.Errorf("response is unsuccessful")
	}
	deposits := make([]Deposit, len(responseStruct.Data))
	for i, datum := range responseStruct.Data {
		deposits[i], err = datum.depositRaw.convert()
		if err != nil {
			return nil, err
		}
	}
	return deposits, err
}

type response struct {
	Data []struct {
		Common
		depositRaw `json:"result"`
	} `json:"data"`
	Success bool `json:"success"`
	Meta    struct {
		At       int64 `json:"at"`
		PageSize int   `json:"page_size"`
	} `json:"meta"`
}
