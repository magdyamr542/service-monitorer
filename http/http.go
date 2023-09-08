package http

import (
	"io"
	gohttp "net/http"
)

type Client interface {
	Get(url string, headers map[string]string) ([]byte, int, error)
}

type http struct {
	client *gohttp.Client
}

func NewClient() Client {
	client := &gohttp.Client{}
	return http{client: client}
}

func (h http) Get(url string, headers map[string]string) ([]byte, int, error) {
	request, err := gohttp.NewRequest("GET", url, nil)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}
