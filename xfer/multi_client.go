package xfer

import (
	"log"
	"sync"

	"github.com/weaveworks/scope/report"
)

// ClientFactory is a thing thats makes AppClients
type ClientFactory func(ProbeConfig, string, string) (AppClient, error)

type multiClient struct {
	ProbeConfig

	clientFactory ClientFactory
	handler       ControlHandler

	mtx     sync.Mutex
	sema    semaphore
	clients map[string]AppClient     // holds map from app id -> client
	ids     map[string]report.IDList // holds map from hostname -> app ids
	quit    chan struct{}
}

type clientTuple struct {
	Details
	AppClient
}

// MultiAppClient maintains a set of upstream apps, and ensures we have an
// AppClient for each one.
type MultiAppClient interface {
	Set(hostname string, endpoints []string)
	Stop()
}

// NewMultiAppClient creates a new MultiAppClient.
func NewMultiAppClient(pc ProbeConfig, handler ControlHandler, clientFactory ClientFactory) MultiAppClient {
	return &multiClient{
		ProbeConfig:   pc,
		clientFactory: clientFactory,
		handler:       handler,

		sema:    newSemaphore(maxConcurrentGET),
		clients: map[string]AppClient{},
		ids:     map[string]report.IDList{},
		quit:    make(chan struct{}),
	}
}

// Set the list of endpoints for the given hostname.
func (c *multiClient) Set(hostname string, endpoints []string) {
	wg := sync.WaitGroup{}
	wg.Add(len(endpoints))
	clients := make(chan clientTuple, len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			c.sema.acquire()
			defer c.sema.release()

			client, err := c.clientFactory(c.ProbeConfig, hostname, endpoint)
			if err != nil {
				log.Printf("Error creating new app client: %v", err)
				return
			}

			details, err := client.Details()
			if err != nil {
				log.Printf("Error fetching app details: %v", err)
			}

			clients <- clientTuple{details, client}
			wg.Done()
		}(endpoint)
	}

	wg.Wait()
	close(clients)
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Start any new apps, and replace the list of app ids for this hostname
	hostIDs := report.MakeIDList()
	for tuple := range clients {
		hostIDs = hostIDs.Add(tuple.ID)

		_, ok := c.clients[tuple.ID]
		if !ok {
			c.clients[tuple.ID] = tuple.AppClient
			tuple.AppClient.ControlConnection(c.handler)
		}
	}
	c.ids[hostname] = hostIDs

	// Remove apps that are no longer referenced (by id) from any hostname
	allReferencedIDs := report.MakeIDList()
	for _, ids := range c.ids {
		allReferencedIDs = allReferencedIDs.Add(ids...)
	}
	for id, client := range c.clients {
		if !allReferencedIDs.Contains(id) {
			client.Stop()
			delete(c.clients, id)
		}
	}
}

// Stop the MultiAppClient.
func (c *multiClient) Stop() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, c := range c.clients {
		c.Stop()
	}
	c.clients = map[string]AppClient{}
	close(c.quit)
}
