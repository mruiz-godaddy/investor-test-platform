package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// forwardRequest performs an HTTP GET to upstream, forwarding auth headers verbatim.
func forwardRequest(upstreamURL, path string, originalReq *http.Request) (*http.Response, error) {
	fullURL := upstreamURL + path
	if originalReq.URL.RawQuery != "" {
		fullURL += "?" + originalReq.URL.RawQuery
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Forward auth headers verbatim
	if auth := originalReq.Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if cookie := originalReq.Header.Get("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	return client.Do(req)
}

// fetchUpstreamListings forwards request to upstream and parses {"listings": [...]} wrapper.
func fetchUpstreamListings(upstreamURL, path string, req *http.Request) ([]map[string]interface{}, error) {
	resp, err := forwardRequest(upstreamURL, path, req)
	if err != nil {
		log.Printf("Upstream %s%s failed: %v", upstreamURL, path, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Upstream %s%s returned status %d", upstreamURL, path, resp.StatusCode)
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Listings []map[string]interface{} `json:"listings"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Listings, nil
}

// fetchUpstreamSearchResults forwards request to upstream and parses {"results": [...]} wrapper.
func fetchUpstreamSearchResults(findUpstream string, req *http.Request) ([]map[string]interface{}, error) {
	resp, err := forwardRequest(findUpstream, req.URL.Path, req)
	if err != nil {
		log.Printf("Find upstream %s%s failed: %v", findUpstream, req.URL.Path, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Find upstream %s%s returned status %d", findUpstream, req.URL.Path, resp.StatusCode)
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Results, nil
}

// mergeListings prepends mock listings, deduplicating by listingId (mock wins).
func mergeListings(mock, upstream []map[string]interface{}) []map[string]interface{} {
	if len(upstream) == 0 {
		return mock
	}
	if len(mock) == 0 {
		return upstream
	}

	// Build set of mock listing IDs
	seen := make(map[interface{}]bool)
	for _, m := range mock {
		if id, ok := m["listingId"]; ok {
			seen[id] = true
		}
		// Also check snake_case variant for search results
		if id, ok := m["auction_id"]; ok {
			seen[id] = true
		}
	}

	result := make([]map[string]interface{}, 0, len(mock)+len(upstream))
	result = append(result, mock...)

	for _, u := range upstream {
		id, ok := u["listingId"]
		if !ok {
			id, ok = u["auction_id"]
		}
		if ok && seen[id] {
			continue // mock wins
		}
		result = append(result, u)
	}

	return result
}
