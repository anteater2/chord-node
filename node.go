package main

import (
	"math/rand"

	"github.com/anteater2/bitmesh/rpc"
	"github.com/anteater2/chord-node/config"
	"github.com/anteater2/chord-node/key"
	"github.com/anteater2/chord-node/node"
	"math"
)

var Address string
var Key key.Key
//var Fingers = make([]*RemoteNode, config.NumFingers())

var RPCCaller *rpc.Caller
var RPCCallee *rpc.Callee

var RPCFindSuccessor rpc.RemoteFunc
var RPCNotify rpc.RemoteFunc

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Fingers []*RemoteNode
	Predecessor *RemoteNode
	Successor *RemoteNode
	Address string
	Key     key.Key
	IsNil   bool
}

// ClosestPrecedingNode finds the closest preceding node to the key in this node's finger table.
// This doesn't need any RPC.
func (node RemoteNode)ClosestPrecedingNode(key key.Key) RemoteNode {
	for i := config.NumFingers() - 1; i >= 0; i-- {
		if node.Fingers[i].IsNil {
			panic("You attempted to find closestPrecedingNode without an initialized finger table!")
		}
		if key.Between(Key, key) {
			return *node.Fingers[i]
		}
	}
	return RemoteNode{Address: Address, Key: Key, IsNil: false}
}

// FindSuccessor finds the successor node to the key.  This may require RPC calls.
func (node RemoteNode)FindSuccessor(key key.Key) RemoteNode {
	if key.Between(Key, node.Successor.Key) {
		// key is between this node and its successor
		return RemoteNode{Address: Address, Key: Key, IsNil: false}
	}

	target := node.ClosestPrecedingNode(key)
	// Now, we have to do an RPC on target to find the successor.
	interf, err := RPCFindSuccessor(target.Address+":2000", uint64(key))

	if err != nil {
		panic("RPC failed")
	}

	rv := interf.(RemoteNode)
	return rv
}

// Notify notifies the successor that you are the predecessor
func (node RemoteNode)Notify(succ RemoteNode) int {
	if succ.Predecessor == nil {
		succ.Predecessor = &node
	}else if succ.Key.Between(node.Predecessor.Key,node.Key){
		succ.Predecessor = &node
	}
	return 0 // Necessary to interface with RPCCaller
}

func (node1 RemoteNode)join(node2 RemoteNode) int{
	node1.Predecessor=nil
	key := node1.Key
	node := node2.FindSuccessor(key)
	node1.Successor = &node
	return 0
}

func (node RemoteNode)CreateLocalNode() {
	// Set the variables of this node.
	Key = key.Key(rand.Uint64() % config.MaxKey()) // Use a random key for now because addresses are all the same rn
	node.Fingers=make([]*RemoteNode,config.NumFingers())
	RPCCaller, err := rpc.NewCaller(config.CallerPort())
	if err != nil {
		panic("RPCCaller failed to initialize")
	}

	RPCCallee, _ := rpc.NewCallee(config.CalleePort())
	if err != nil {
		panic("RPCCallee failed to initialize")
	}

	Address = "127.0.0.1" // TODO: Make the address resolve to the real address of the node.
	node.Predecessor = nil
	node.Successor = &RemoteNode{
		Address: "127.0.0.1", // TODO: Make the address resolve to the real address of the node.
		Key:     Key,
		IsNil:   false,
	}

	// Define all of the RPC functions.  For more info, look at Yuchen's caller.go and example_test.go
	RPCFindSuccessor = RPCCaller.Declare(key.NewKey(1), RemoteNode{}, 10)
	RPCNotify = RPCCaller.Declare(RemoteNode{}, 0, 10)

	 //RPCIsAlive = RPCCaller.Declare(nil{}, nil{}, 0) // Don't define this just yet - how does the RPC system react if a node fails to respond?
	 //Hook the RPCCallee into this node's functions
	RPCCallee.Implement(node.FindSuccessor) // What happens if two methods have the same arg type signature?
	RPCCallee.Implement(node.Notify)
}

func (node RemoteNode)Stablize(){
	x:=node.Successor.Predecessor
	if x.Key.Between(node.Key,node.Successor.Key){
		node.Successor=x
	}
	node.Successor.Notify(node)
}

func (node RemoteNode)fixfingers(){
	finger_num := config.NumFingers()
	m:=math.Log2(float64(finger_num))
	for i :=0;i<int(m);i++{
		nextkey := uint64(node.Key)+ uint64(math.Exp2(float64(i-1)))
			succ :=node.FindSuccessor(key.NewKey(nextkey))
		node.Fingers[i]= &succ
	}
}
// func main() {
// 	CreateLocalNode()
// 	go RPCCallee.Start()
// 	go RPCCaller.Start()
// 	// add the period functions here
// }
