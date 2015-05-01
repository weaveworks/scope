package router

import (
	"bytes"
	"fmt"
	"math"
	"sync"
)

type Routes struct {
	sync.RWMutex
	ourself      *Peer
	peers        *Peers
	unicast      map[PeerName]PeerName
	unicastAll   map[PeerName]PeerName // [1]
	broadcast    map[PeerName][]PeerName
	broadcastAll map[PeerName][]PeerName // [1]
	recalculate  chan<- *struct{}
	wait         chan<- chan struct{}
	// [1] based on *all* connections, not just established &
	// symmetric ones
}

func NewRoutes(ourself *Peer, peers *Peers) *Routes {
	routes := &Routes{
		ourself:      ourself,
		peers:        peers,
		unicast:      make(map[PeerName]PeerName),
		unicastAll:   make(map[PeerName]PeerName),
		broadcast:    make(map[PeerName][]PeerName),
		broadcastAll: make(map[PeerName][]PeerName)}
	routes.unicast[ourself.Name] = UnknownPeerName
	routes.unicastAll[ourself.Name] = UnknownPeerName
	routes.broadcast[ourself.Name] = []PeerName{}
	routes.broadcastAll[ourself.Name] = []PeerName{}
	return routes
}

func (routes *Routes) Start() {
	recalculate := make(chan *struct{}, 1)
	wait := make(chan chan struct{})
	routes.recalculate = recalculate
	routes.wait = wait
	go routes.run(recalculate, wait)
}

func (routes *Routes) PeerNames() PeerNameSet {
	return routes.peers.Names()
}

func (routes *Routes) Unicast(name PeerName) (PeerName, bool) {
	routes.RLock()
	defer routes.RUnlock()
	hop, found := routes.unicast[name]
	return hop, found
}

func (routes *Routes) UnicastAll(name PeerName) (PeerName, bool) {
	routes.RLock()
	defer routes.RUnlock()
	hop, found := routes.unicastAll[name]
	return hop, found
}

func (routes *Routes) Broadcast(name PeerName) []PeerName {
	routes.RLock()
	defer routes.RUnlock()
	hops, found := routes.broadcast[name]
	if !found {
		return []PeerName{}
	}
	return hops
}

func (routes *Routes) BroadcastAll(name PeerName) []PeerName {
	routes.RLock()
	defer routes.RUnlock()
	hops, found := routes.broadcastAll[name]
	if !found {
		return []PeerName{}
	}
	return hops
}

// Choose min(log2(n_peers), n_neighbouring_peers) neighbours, with a
// random distribution that is topology-sensitive, favouring
// neighbours at the end of "bottleneck links". We determine the
// latter based on the unicast routing table. If a neighbour appears
// as the value more frequently than others - meaning that we reach a
// higher proportion of peers via that neighbour than other neighbours
// - then it is chosen with a higher probability.
//
// Note that we choose log2(n_peers) *neighbours*, not
// peers. Consequently, on sparsely connected peers this function
// returns a higher proportion of neighbours than elsewhere. In
// extremis, on peers with fewer than log2(n_peers) neighbours, all
// neighbours are returned.
func (routes *Routes) RandomNeighbours(except PeerName) PeerNameSet {
	res := make(PeerNameSet)
	routes.RLock()
	defer routes.RUnlock()
	count := int(math.Log2(float64(len(routes.unicastAll))))
	// depends on go's random map iteration
	for _, dst := range routes.unicastAll {
		if dst != UnknownPeerName && dst != except {
			res[dst] = void
			if len(res) >= count {
				break
			}
		}
	}
	return res
}

func (routes *Routes) String() string {
	var buf bytes.Buffer
	routes.RLock()
	defer routes.RUnlock()
	fmt.Fprintln(&buf, "unicast:")
	for name, hop := range routes.unicast {
		fmt.Fprintf(&buf, "%s -> %s\n", name, hop)
	}
	fmt.Fprintln(&buf, "broadcast:")
	for name, hops := range routes.broadcast {
		fmt.Fprintf(&buf, "%s -> %v\n", name, hops)
	}
	// We don't include the 'all' routes here since they are of
	// limited utility in troubleshooting
	return buf.String()
}

// Request recalculation of the routing table. This is async but can
// effectively be made synchronous with a subsequent call to
// EnsureRecalculated.
func (routes *Routes) Recalculate() {
	// The use of a 1-capacity channel in combination with the
	// non-blocking send is an optimisation that results in multiple
	// requests being coalesced.
	select {
	case routes.recalculate <- nil:
	default:
	}
}

// Wait for any preceding Recalculate requests to be processed.
func (routes *Routes) EnsureRecalculated() {
	done := make(chan struct{})
	routes.wait <- done
	<-done
}

func (routes *Routes) run(recalculate <-chan *struct{}, wait <-chan chan struct{}) {
	for {
		select {
		case <-recalculate:
			routes.calculate()
		case done := <-wait:
			select {
			case <-recalculate:
				routes.calculate()
			default:
			}
			close(done)
		}
	}
}

func (routes *Routes) calculate() {
	var (
		unicast      = routes.calculateUnicast(true)
		unicastAll   = routes.calculateUnicast(false)
		broadcast    = routes.calculateBroadcast(true)
		broadcastAll = routes.calculateBroadcast(false)
	)
	routes.Lock()
	routes.unicast = unicast
	routes.unicastAll = unicastAll
	routes.broadcast = broadcast
	routes.broadcastAll = broadcastAll
	routes.Unlock()
}

// Calculate all the routes for the question: if *we* want to send a
// packet to Peer X, what is the next hop?
//
// When we sniff a packet, we determine the destination peer
// ourself. Consequently, we can relay the packet via any
// arbitrary peers - the intermediate peers do not have to have
// any knowledge of the MAC address at all. Thus there's no need
// to exchange knowledge of MAC addresses, nor any constraints on
// the routes that we construct.
func (routes *Routes) calculateUnicast(establishedAndSymmetric bool) map[PeerName]PeerName {
	_, unicast := routes.ourself.Routes(nil, establishedAndSymmetric)
	return unicast
}

// Calculate all the routes for the question: if we receive a
// broadcast originally from Peer X, which peers should we pass the
// frames on to?
//
// When the topology is stable, and thus all peers perform route
// calculations based on the same data, the algorithm ensures that
// broadcasts reach every peer exactly once.
//
// This is largely due to properties of the Peer.Routes algorithm. In
// particular:
//
// ForAll X,Y,Z in Peers.
//     X.Routes(Y) <= X.Routes(Z) \/
//     X.Routes(Z) <= X.Routes(Y)
// ForAll X,Y,Z in Peers.
//     Y =/= Z /\ X.Routes(Y) <= X.Routes(Z) =>
//     X.Routes(Y) u [P | Y.HasSymmetricConnectionTo(P)] <= X.Routes(Z)
// where <= is the subset relationship on keys of the returned map.
func (routes *Routes) calculateBroadcast(establishedAndSymmetric bool) map[PeerName][]PeerName {
	broadcast := make(map[PeerName][]PeerName)
	ourself := routes.ourself
	ourConnections := ourself.Connections()

	routes.peers.ForEach(func(peer *Peer) {
		hops := []PeerName{}
		if found, reached := peer.Routes(ourself, establishedAndSymmetric); found {
			// This is rather similar to the inner loop on
			// peer.Routes(...); the main difference is in the
			// locking.
			for conn := range ourConnections {
				if establishedAndSymmetric && !conn.Established() {
					continue
				}
				remoteName := conn.Remote().Name
				if _, found := reached[remoteName]; found {
					continue
				}
				if remoteConn, found := conn.Remote().ConnectionTo(ourself.Name); !establishedAndSymmetric || (found && remoteConn.Established()) {
					hops = append(hops, remoteName)
				}
			}
		}
		broadcast[peer.Name] = hops
	})
	return broadcast
}
