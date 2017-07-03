package kubernetes

import (
	"sync"

	"k8s.io/client-go/tools/cache"
)

// Event type is an enum of ADD, UPDATE and DELETE
type Event int

// Watch type is for callbacks when somethings happens to the store.
type Watch func(Event, interface{})

// Event enum values.
const (
	ADD Event = iota
	UPDATE
	DELETE
)

type eventStore struct {
	mtx     sync.Mutex
	watch   Watch
	keyFunc cache.KeyFunc
	cache.Store
}

// NewEventStore creates a new Store which triggers watch whenever
// an object is added, removed or updated.
func NewEventStore(watch Watch, keyFunc cache.KeyFunc) cache.Store {
	return &eventStore{
		keyFunc: keyFunc,
		watch:   watch,
		Store:   cache.NewStore(keyFunc),
	}
}

func (e *eventStore) Add(o interface{}) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.watch(ADD, o)
	return e.Store.Add(o)
}

func (e *eventStore) Update(o interface{}) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.watch(UPDATE, o)
	return e.Store.Update(o)
}

func (e *eventStore) Delete(o interface{}) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.watch(DELETE, o)
	return e.Store.Delete(o)
}

func (e *eventStore) Replace(os []interface{}, ver string) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	indexed := map[string]interface{}{}
	for _, o := range os {
		key, err := e.keyFunc(o)
		if err != nil {
			return err
		}
		indexed[key] = o
	}

	existing := map[string]interface{}{}
	for _, o := range e.Store.List() {
		key, err := e.keyFunc(o)
		if err != nil {
			return err
		}
		existing[key] = o
		if _, ok := indexed[key]; !ok {
			e.watch(DELETE, o)
		}
	}

	for key, o := range indexed {
		if _, ok := existing[key]; !ok {
			e.watch(ADD, o)
		} else {
			e.watch(UPDATE, o)
		}
	}

	return e.Store.Replace(os, ver)
}
