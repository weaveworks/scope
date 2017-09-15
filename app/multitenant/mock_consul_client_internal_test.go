package multitenant

import (
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type mockKV struct {
	mtx     sync.Mutex
	cond    *sync.Cond
	kvps    map[string]*consul.KVPair
	next    uint64 // the next update will have this 'index in the the log'
	quit    chan struct{}
	stopped bool
}

func newMockConsulClient(quit chan struct{}) ConsulClient {
	m := mockKV{
		kvps: map[string]*consul.KVPair{},
		quit: quit,
	}
	m.cond = sync.NewCond(&m.mtx)
	go m.loop()
	return &consulClient{&m}
}

func copyKVPair(in *consul.KVPair) *consul.KVPair {
	value := make([]byte, len(in.Value))
	copy(value, in.Value)
	return &consul.KVPair{
		Key:         in.Key,
		CreateIndex: in.CreateIndex,
		ModifyIndex: in.ModifyIndex,
		LockIndex:   in.LockIndex,
		Flags:       in.Flags,
		Value:       value,
		Session:     in.Session,
	}
}

// periodic loop to wake people up, so they can honour timeouts
func (m *mockKV) loop() {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			m.mtx.Lock()
			m.cond.Broadcast()
			m.mtx.Unlock()
		case <-m.quit:
			m.stopped = true
			return
		}
	}
}

func (m *mockKV) CAS(p *consul.KVPair, q *consul.WriteOptions) (bool, *consul.WriteMeta, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	existing, ok := m.kvps[p.Key]
	if ok && existing.ModifyIndex != p.ModifyIndex {
		return false, nil, nil
	}
	if ok {
		existing.Value = p.Value
	} else {
		m.kvps[p.Key] = copyKVPair(p)
	}
	m.kvps[p.Key].ModifyIndex++
	m.kvps[p.Key].LockIndex = m.next
	m.next++
	m.cond.Broadcast()
	return true, nil, nil
}

func (m *mockKV) Get(key string, q *consul.QueryOptions) (*consul.KVPair, *consul.QueryMeta, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	value, ok := m.kvps[key]
	if !ok {
		return nil, nil, nil
	}
	for q.WaitIndex >= value.ModifyIndex && !m.stopped {
		m.cond.Wait()
	}
	return copyKVPair(value), nil, nil
}

func (m *mockKV) List(prefix string, q *consul.QueryOptions) (consul.KVPairs, *consul.QueryMeta, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	deadline := time.Now().Add(q.WaitTime)
	for m.next <= q.WaitIndex && time.Now().Before(deadline) && !m.stopped {
		m.cond.Wait()
	}
	if time.Now().After(deadline) {
		return nil, &consul.QueryMeta{LastIndex: q.WaitIndex}, nil
	}
	result := consul.KVPairs{}
	for _, kvp := range m.kvps {
		if kvp.LockIndex >= q.WaitIndex {
			result = append(result, copyKVPair(kvp))
		}
	}
	return result, &consul.QueryMeta{LastIndex: m.next}, nil
}
