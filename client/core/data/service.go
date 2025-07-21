package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DataService struct{}

func NewDataService() *DataService {
	return &DataService{}
}

// makeRequest constructs and executes an HTTP request with dployr server
func (d *DataService) makeRequest(method, endpoint, host, token string, queryParams map[string]string, body interface{}) (*http.Response, error) {
	urlStr := fmt.Sprintf("http://%s:7879/v1/%s", host, endpoint)
	
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if queryParams != nil {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, errBody)
	}

	return resp, nil
}

// decodeResponse decodes the HTTP response body into the target interface
func (d *DataService) decodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

