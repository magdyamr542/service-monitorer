package http

import (
	"io"
	gohttp "net/http"
)

type Client interface {
	Get(url string, headers map[string]string) ([]byte, int, error)
	Post(url string, headers map[string]string, body io.Reader) ([]byte, int, error)
}

type http struct {
	client *gohttp.Client
}

func NewClient() Client {
	client := &gohttp.Client{}
	return http{client: client}
}

func (h http) Get(url string, headers map[string]string) ([]byte, int, error) {
	return h.request("GET", url, headers, nil)
}

func (h http) Post(url string, headers map[string]string, body io.Reader) ([]byte, int, error) {
	return h.request("POST", url, headers, body)
}

func (h http) request(method, url string, headers map[string]string, body io.Reader) ([]byte, int, error) {
	request, err := gohttp.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	resp, err := h.client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return responseBody, resp.StatusCode, nil
}
