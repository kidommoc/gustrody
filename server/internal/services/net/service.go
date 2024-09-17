package net

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
)

const CLIENT_NUM = 10
const HTTP_TIMEOUT = time.Minute

type NetService struct {
	lg         logging.Logger
	clientPool chan *http.Client
}

func NewNetService(lg logging.Logger) *NetService {
	service := &NetService{lg: lg, clientPool: make(chan *http.Client, CLIENT_NUM)}
	for i := 0; i < CLIENT_NUM; i++ {
		service.clientPool <- &http.Client{Timeout: HTTP_TIMEOUT}
	}
	return service
}

type Client struct {
	lg      logging.Logger
	client  *http.Client
	service *NetService
}

// block when no client available
func (service *NetService) HttpClient() *Client {
	return &Client{
		lg:      service.lg,
		client:  <-service.clientPool,
		service: service,
	}
}

func (client *Client) Close() {
	if client.client == nil {
		return
	}
	client.service.clientPool <- client.client
	client.client = nil
	client.service = nil
}

func (client *Client) Do(req *http.Request, body []byte) (response *Response, err error) {
	logger := client.lg
	if body != nil {
		b := io.NopCloser(bytes.NewBuffer(body))
		req.Body = b
	}
	res, err := client.client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("[Net] Failed to send request to %s.", req.URL)
		logger.Error(msg, err)
		return nil, err
	}
	if res.StatusCode > http.StatusBadRequest {
		err := fmt.Errorf("status code: %d", res.StatusCode)
		msg := fmt.Sprintf("[Net] Response not ok of request to %s.", req.URL)
		logger.Error(msg, err)
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		msg := fmt.Sprintf("[Net] Failed to read response body of request to %s.", req.URL)
		logger.Error(msg, err)
		return nil, err
	}
	return &Response{
		Header: res.Header,
		Body:   resBody,
	}, nil
}

type Response struct {
	Header http.Header
	Body   []byte
}
