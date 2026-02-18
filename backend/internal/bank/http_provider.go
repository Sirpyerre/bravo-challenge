package bank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type httpProvider struct {
	country string
	baseURL string
	client  *http.Client
}

func newHTTPProvider(country, baseURL string) BankProvider {
	return &httpProvider{
		country: country,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *httpProvider) Country() string { 
	return p.country 
}

func (p *httpProvider) Evaluate(ctx context.Context, req EvaluationRequest) (*EvaluationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/evaluate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call bank %s: %w", p.country, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bank %s returned status %d", p.country, resp.StatusCode)
	}

	var evalResp EvaluationResponse
	if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
		return nil, fmt.Errorf("decode bank response: %w", err)
	}

	return &evalResp, nil
}
