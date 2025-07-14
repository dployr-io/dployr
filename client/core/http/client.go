// http/client.go
package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Get(url string) (any, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) Post(url string, data interface{}) (any, error) {
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)
	return responseData, err
}

func (c *Client) Put(url string, data interface{}) (any, error) {
	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)
	return responseData, err
}

func (c *Client) Delete(url string) (any, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)
	return responseData, err
}
