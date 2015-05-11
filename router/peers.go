package router

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
)

type Peers struct {
	sync.RWMutex
	ourself *Peer
	table   map[PeerName]*Peer
	onGC    func(*Peer)
}

type UnknownPeerError struct {
	Name PeerName
}

type NameCollisionError struct {
	Name PeerName
}

type PeerNameSet map[PeerName]struct{}

type PeerSummary struct {
	NameByte []byte
	NickName string
	UID      uint64
	Version  uint64
}

type ConnectionSummary struct {
	NameByte      []byte
	RemoteTCPAddr string
	Outbound      bool
	Established   bool
}

func NewPeers(ourself *Peer, onGC func(*Peer)) *Peers {
	return &Peers{
		ourself: ourself,
		table:   make(map[PeerName]*Peer),
		onGC:    onGC}
}

func (peers *Peers) FetchWithDefault(peer *Peer) *Peer {
	peers.RLock()
	res, found := peers.fetchAlias(peer)
	peers.RUnlock()
	if found {
		return res
	}
	peers.Lock()
	defer peers.Unlock()
	res, found = peers.fetchAlias(peer)
	if found {
		return res
	}
	peers.table[peer.Name] = peer
	peer.IncrementLocalRefCount()
	return peer
}

func (peers *Peers) Fetch(name PeerName) (*Peer, bool) {
	peers.RLock()
	defer peers.RUnlock()
	peer, found := peers.table[name]
	return peer, found // GRRR, why can't I inline this!?
}

func (peers *Peers) ForEach(fun func(*Peer)) {
	peers.RLock()
	defer peers.RUnlock()
	for _, peer := range peers.table {
		fun(peer)
	}
}

// Merge an incoming update with our own topology.
//
// We add peers hitherto unknown to us, and update peers for which the
// update contains a more recent version than known to us. The return
// value is a) a representation of the received update, and b) an
// "improved" update containing just these new/updated elements.
func (peers *Peers) ApplyUpdate(update []byte) (PeerNameSet, PeerNameSet, error) {
	peers.Lock()

	newPeers, decodedUpdate, decodedConns, err := peers.decodeUpdate(update)
	if err != nil {
		peers.Unlock()
		return nil, nil, err
	}

	// By this point, we know the update doesn't refer to any peers we
	// have no knowledge of. We can now apply the update. Start by
	// adding in any new peers into the cache.
	for name, newPeer := range newPeers {
		peers.table[name] = newPeer
	}

	// Now apply the updates
	newUpdate := peers.applyUpdate(decodedUpdate, decodedConns)

	for _, peerRemoved := range peers.garbageCollect() {
		delete(newUpdate, peerRemoved.Name)
	}

	// Don't need to hold peers lock any longer
	peers.Unlock()

	updateNames := make(PeerNameSet)
	for _, peer := range decodedUpdate {
		updateNames[peer.Name] = void
	}
	return updateNames, setFromPeersMap(newUpdate), nil
}

func (peers *Peers) Names() PeerNameSet {
	peers.RLock()
	defer peers.RUnlock()
	return setFromPeersMap(peers.table)
}

func (peers *Peers) EncodePeers(names PeerNameSet) []byte {
	peers.RLock()
	peerList := make([]*Peer, 0, len(names))
	for name := range names {
		if peer, found := peers.table[name]; found {
			peerList = append(peerList, peer)
		}
	}
	peers.RUnlock() // release lock so we don't hold it while encoding
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	for _, peer := range peerList {
		peer.encode(enc)
	}
	return buf.Bytes()
}

func (peers *Peers) GarbageCollect() []*Peer {
	peers.Lock()
	defer peers.Unlock()
	return peers.garbageCollect()
}

func (peers *Peers) String() string {
	var buf bytes.Buffer
	peers.ForEach(func(peer *Peer) {
		fmt.Fprintln(&buf, peer.Info())
		for conn := range peer.Connections() {
			established := ""
			if !conn.Established() {
				established = " (unestablished)"
			}
			fmt.Fprintf(&buf, "   -> %s [%v%s]\n", conn.Remote(), conn.RemoteTCPAddr(), established)
		}
	})
	return buf.String()
}

func (peers *Peers) fetchAlias(peer *Peer) (*Peer, bool) {
	if existingPeer, found := peers.table[peer.Name]; found {
		if existingPeer.UID != peer.UID {
			return nil, true
		}
		existingPeer.IncrementLocalRefCount()
		return existingPeer, true
	}
	return nil, false
}

func (peers *Peers) garbageCollect() []*Peer {
	removed := []*Peer{}
	_, reached := peers.ourself.Routes(nil, false)
	for name, peer := range peers.table {
		if _, found := reached[peer.Name]; !found && !peer.IsLocallyReferenced() {
			delete(peers.table, name)
			peers.onGC(peer)
			removed = append(removed, peer)
		}
	}
	return removed
}

func setFromPeersMap(peers map[PeerName]*Peer) PeerNameSet {
	names := make(PeerNameSet)
	for name := range peers {
		names[name] = void
	}
	return names
}

func (peers *Peers) decodeUpdate(update []byte) (newPeers map[PeerName]*Peer, decodedUpdate []*Peer, decodedConns [][]ConnectionSummary, err error) {
	newPeers = make(map[PeerName]*Peer)
	decodedUpdate = []*Peer{}
	decodedConns = [][]ConnectionSummary{}

	updateBuf := new(bytes.Buffer)
	updateBuf.Write(update)
	decoder := gob.NewDecoder(updateBuf)

	for {
		peerSummary, connSummaries, decErr := decodePeer(decoder)
		if decErr == io.EOF {
			break
		} else if decErr != nil {
			err = decErr
			return
		}
		name := PeerNameFromBin(peerSummary.NameByte)
		newPeer := NewPeer(name, peerSummary.NickName, peerSummary.UID, peerSummary.Version)
		decodedUpdate = append(decodedUpdate, newPeer)
		decodedConns = append(decodedConns, connSummaries)
		existingPeer, found := peers.table[name]
		if !found {
			newPeers[name] = newPeer
		} else if existingPeer.UID != newPeer.UID {
			err = NameCollisionError{Name: newPeer.Name}
			return
		}
	}

	for _, connSummaries := range decodedConns {
		for _, connSummary := range connSummaries {
			remoteName := PeerNameFromBin(connSummary.NameByte)
			if _, found := newPeers[remoteName]; found {
				continue
			}
			if _, found := peers.table[remoteName]; found {
				continue
			}
			// Update refers to a peer which we have no knowledge
			// of. Thus we can't apply the update. Abort.
			err = UnknownPeerError{remoteName}
			return
		}
	}
	return
}

func (peers *Peers) applyUpdate(decodedUpdate []*Peer, decodedConns [][]ConnectionSummary) map[PeerName]*Peer {
	newUpdate := make(map[PeerName]*Peer)
	for idx, newPeer := range decodedUpdate {
		connSummaries := decodedConns[idx]
		name := newPeer.Name
		// guaranteed to find peer in the peers.table
		peer := peers.table[name]
		if peer != newPeer &&
			(peer == peers.ourself || peer.Version() >= newPeer.Version()) {
			// Nobody but us updates us. And if we know more about a
			// peer than what's in the the update, we ignore the
			// latter.
			continue
		}
		// If we're here, either it was a new peer, or the update has
		// more info about the peer than we do. Either case, we need
		// to set version and conns and include the updated peer in
		// the outgoing update.

		// Can peer have been updated by anyone else in the mean time?
		// No - we know that peer is not ourself, so the only prospect
		// for an update would be someone else calling
		// router.Peers.ApplyUpdate. But ApplyUpdate takes the Lock on
		// the router.Peers, so there can be no race here.
		conns := makeConnsMap(peer, connSummaries, peers.table)
		peer.SetVersionAndConnections(newPeer.Version(), conns)
		newUpdate[name] = peer
	}
	return newUpdate
}

func (peer *Peer) encode(enc *gob.Encoder) {
	peer.RLock()
	defer peer.RUnlock()

	checkFatal(enc.Encode(PeerSummary{
		peer.NameByte,
		peer.NickName,
		peer.UID,
		peer.version}))

	connSummaries := []ConnectionSummary{}
	for _, conn := range peer.connections {
		connSummaries = append(connSummaries, ConnectionSummary{
			conn.Remote().NameByte,
			conn.RemoteTCPAddr(),
			conn.Outbound(),
			// DANGER holding rlock on peer, going to take rlock on conn
			conn.Established(),
		})
	}

	checkFatal(enc.Encode(connSummaries))
}

func decodePeer(dec *gob.Decoder) (peerSummary PeerSummary, connSummaries []ConnectionSummary, err error) {
	if err = dec.Decode(&peerSummary); err != nil {
		return
	}
	if err = dec.Decode(&connSummaries); err != nil {
		return
	}
	return
}

func makeConnsMap(peer *Peer, connSummaries []ConnectionSummary, table map[PeerName]*Peer) map[PeerName]Connection {
	conns := make(map[PeerName]Connection)
	for _, connSummary := range connSummaries {
		name := PeerNameFromBin(connSummary.NameByte)
		remotePeer := table[name]
		conn := NewRemoteConnection(peer, remotePeer, connSummary.RemoteTCPAddr, connSummary.Outbound, connSummary.Established)
		conns[name] = conn
	}
	return conns
}
