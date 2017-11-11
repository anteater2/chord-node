package main

import (
	"math/rand"

	"github.com/anteater2/bitmesh/rpc"
	"github.com/anteater2/chord-node/config"
	"github.com/anteater2/chord-node/key"
)

var Address string
var Key key.Key
var Fingers = make([]*RemoteNode, config.NumFingers())
var Predecessor *RemoteNode
var Successor *RemoteNode
var RPCCaller *rpc.Caller
var RPCCallee *rpc.Callee

// This is how you declare a function pointer in go:
// var RPCFindSuccessor func(string, uint32) RemoteNode
// However, we have a custom RemoteFunc type, so we might as well use that.
var RPCFindSuccessor rpc.RemoteFunc
var RPCNotify rpc.RemoteFunc

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Address string
	Key     key.Key
	IsNil   bool
}

// func isInExclusive(key uint32, start uint32, end uint32) bool {
// 	key = key % MaxKey
// 	if start > MaxKey {
// 		panic("MaxKey constraint violated by start")
// 	}
// 	if end > MaxKey {
// 		panic("MaxKey constraint violated by end")
// 	}
// 	if start < end {
// 		return key > start && key < end
// 	}
// 	if start >= end {
// 		return key > start || key < end
// 	}
// 	return false // What a stupid compiler
// }

// ClosestPrecedingNode finds the closest preceding node to the key in this node's finger table.
// This doesn't need any RPC.
func ClosestPrecedingNode(key key.Key) RemoteNode {
	for i := config.NumFingers() - 1; i >= 0; i-- {
		if Fingers[i].IsNil {
			panic("You attempted to find closestPrecedingNode without an initialized finger table!")
		}
		if key.Between(Key, key) {
			return *Fingers[i]
		}
	}
	return RemoteNode{Address: Address, Key: Key, IsNil: false}
}

// FindSuccessor finds the successor node to the key.  This may require RPC calls.
func FindSuccessor(key key.Key) RemoteNode {
	if key.Between(Key, Successor.Key) {
		// key is between this node and its successor
		return RemoteNode{Address: Address, Key: Key, IsNil: false}
	}

	target := ClosestPrecedingNode(key)
	// Now, we have to do an RPC on target to find the successor.
	interf, err := RPCFindSuccessor(target.Address+":2000", uint64(key))

	if err != nil {
		panic("RPC failed")
	}

	rv := interf.(RemoteNode)
	return rv
}

func Notify(node RemoteNode) {
	if Predecessor == nil {
		Predecessor = &node
	}
}

func CreateLocalNode() {
	// Set the variables of this node.
	Key = key.Key(rand.Uint64() % config.MaxKey()) // Use a random key for now because addresses are all the same rn

	RPCCaller, err := rpc.NewCaller(config.CallerPort())
	if err != nil {
		panic("RPCCaller failed to initialize")
	}

	RPCCallee, _ := rpc.NewCallee(config.CalleePort())
	if err != nil {
		panic("RPCCallee failed to initialize")
	}

	Address = "127.0.0.1" // TODO: Make the address resolve to the real address of the node.
	Predecessor = nil
	Successor = &RemoteNode{
		Address: "127.0.0.1", // TODO: Make the address resolve to the real address of the node.
		Key:     Key,
		IsNil:   false,
	}

	// Define all of the RPC functions.  For more info, look at Yuchen's caller.go and example_test.go
	RPCFindSuccessor = RPCCaller.Declare(key.NewKey(1), RemoteNode{}, 10)
	RPCNotify = RPCCaller.Declare(RemoteNode{}, 0, 10)

	// RPCIsAlive = RPCCaller.Declare(nil{}, nil{}, 0) // Don't define this just yet - how does the RPC system react if a node fails to respond?
	// Hook the RPCCallee into this node's functions
	RPCCallee.Implement(FindSuccessor) // What happens if two methods have the same arg type signature?
	RPCCallee.Implement(Notify)

}

func main() {
	CreateLocalNode()
	go RPCCallee.Start()
	go RPCCaller.Start()
	// add the period functions here
}
