package net

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/kidommoc/gustrody/internal/logging"
)

const clientNum = 10

type NetService struct {
	lg         logging.Logger
	mutex      sync.Mutex
	lsnHead    *lnode // listener, queue head
	lsnTail    *lnode // listener, queue tail
	clientPool *cnode
}

func NewNetService(lg logging.Logger) *NetService {
	service := &NetService{lg: lg}
	for i := 0; i < clientNum; i += 1 {
		node := &cnode{next: service.clientPool, client: http.DefaultClient}
		service.clientPool = node
	}
	return service
}

type lnode struct {
	before *lnode
	next   *lnode
	ch     chan bool
}

type cnode struct {
	next   *cnode
	client *http.Client
}

type Client struct {
	lg      logging.Logger
	client  *http.Client
	service *NetService
}

// block when no client available
func (service *NetService) HttpClient() *Client {
	service.mutex.Lock()
	for service.clientPool == nil {
		node := lnode{ch: make(chan bool)}
		if service.lsnHead == nil {
			service.lsnHead = &node
			service.lsnTail = &node
		} else {
			node.before = service.lsnTail
			service.lsnTail.next = &node
			service.lsnTail = &node
		}
		service.mutex.Unlock()
		<-node.ch
	}

	client := Client{
		lg:      service.lg,
		client:  service.clientPool.client,
		service: service,
	}
	service.clientPool = service.clientPool.next
	service.mutex.Unlock()
	return &client
}

func (client *Client) Close() {
	if client.client == nil {
		return
	}
	client.service.mutex.Lock()
	node := cnode{
		next:   client.service.clientPool,
		client: client.client,
	}
	client.service.clientPool = &node

	if client.service.lsnHead != nil {
		sig := client.service.lsnHead.ch
		if client.service.lsnHead.next != nil {
			client.service.lsnHead.next.before = nil
		}
		client.service.lsnHead = client.service.lsnHead.next
		sig <- true
	} else {
		client.service.mutex.Unlock()
	}

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
		logger.Error("[Net] Failed to send request.", err)
		return nil, err
	}
	if res.StatusCode > http.StatusBadRequest {
		err := fmt.Errorf("status code: %d", res.StatusCode)
		logger.Error("[Net] Response not ok.", err)
		return nil, err
	}
	var resBody []byte
	if _, err = res.Body.Read(resBody); err != nil {
		logger.Error("[Net] Failed to read response body.", err)
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
