package net

import (
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

func (client *Client) Do(req *http.Request) (res *http.Response, err error) {
	return client.client.Do(req)
}
