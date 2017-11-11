package node

import (
	"math/rand"
	"net"

	"github.com/anteater2/chord-node/config"
	"github.com/anteater2/chord-node/key"
)

// Node is a node on the chord network
type Node struct {
	Address      net.Addr
	Key          key.Key
	Successors   []*Node
	Predecessors []*Node
	Fingers      []*Node
}

var initialized = false

var localNode = Node{}

// ClosestNode returns the closest (preceding) node to the key
func (node Node) ClosestNode(key key.Key) Node {
	for i := config.NumFingers() - 1; i >= 0; i-- {
		if node.Fingers[i] == nil {
			panic("finger table is uninitialized")
		}
		if key.Between(node.Key, key) {
			return *node.Fingers[i]
		}
	}

	return node
}

// connect connects the local node to the network given an address
func connect(addr string) {
	// a bunch of logic here
	localNode.Successors = make([]*Node, 5)
	localNode.Predecessors = make([]*Node, 5)
	localNode.Key = key.Key(rand.Uint64() % config.MaxKey())
}

// LocalNode returns the local node singleton
func LocalNode() Node {
	if !initialized {
		if config.Creator() {
			initNetwork()
		} else {
			connect(config.Introducer())
		}
		initialized = true
	}
	return localNode
}

// initNetwork initializes the node as the first node in a new chord network
func initNetwork() {
	localNode = Node{
		Address:      config.Addr(),
		Successors:   make([]*Node, 3),
		Predecessors: make([]*Node, 3),
		Fingers:      make([]*Node, config.NumFingers()),
	}
}
