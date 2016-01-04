package appclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/report"
)

const maxConcurrentGET = 10

// ClientFactory is a thing thats makes AppClients
type ClientFactory func(string, string) (AppClient, error)

type multiClient struct {
	clientFactory ClientFactory

	mtx     sync.Mutex
	sema    semaphore
	clients map[string]AppClient     // holds map from app id -> client
	ids     map[string]report.IDList // holds map from hostname -> app ids
	quit    chan struct{}
}

type clientTuple struct {
	xfer.Details
	AppClient
}

// Publisher is something which can send a stream of data somewhere, probably
// to a remote collector.
type Publisher interface {
	Publish(io.Reader) error
	Stop()
}

// MultiAppClient maintains a set of upstream apps, and ensures we have an
// AppClient for each one.
type MultiAppClient interface {
	Set(hostname string, endpoints []string)
	PipeConnection(appID, pipeID string, pipe xfer.Pipe) error
	PipeClose(appID, pipeID string) error
	Stop()
	Publish(io.Reader) error
}

// NewMultiAppClient creates a new MultiAppClient.
func NewMultiAppClient(clientFactory ClientFactory) MultiAppClient {
	return &multiClient{
		clientFactory: clientFactory,

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

			client, err := c.clientFactory(hostname, endpoint)
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
			tuple.AppClient.ControlConnection()
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

func (c *multiClient) withClient(appID string, f func(AppClient) error) error {
	c.mtx.Lock()
	client, ok := c.clients[appID]
	c.mtx.Unlock()
	if !ok {
		return fmt.Errorf("No such app id: %s", appID)
	}
	return f(client)
}

func (c *multiClient) PipeConnection(appID, pipeID string, pipe xfer.Pipe) error {
	return c.withClient(appID, func(client AppClient) error {
		client.PipeConnection(pipeID, pipe)
		return nil
	})
}

func (c *multiClient) PipeClose(appID, pipeID string) error {
	return c.withClient(appID, func(client AppClient) error {
		return client.PipeClose(pipeID)
	})
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

// Publish implements Publisher by publishing the reader to all of the
// underlying publishers sequentially. To do that, it needs to drain the
// reader, and recreate new readers for each publisher. Note that it will
// publish to one endpoint for each unique ID. Failed publishes don't count.
func (c *multiClient) Publish(r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	errs := []string{}
	for _, c := range c.clients {
		if err := c.Publish(bytes.NewReader(buf)); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type semaphore chan struct{}

func newSemaphore(n int) semaphore {
	c := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}
	return semaphore(c)
}
func (s semaphore) acquire() { <-s }
func (s semaphore) release() { s <- struct{}{} }
