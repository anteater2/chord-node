package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"./key"
	"github.com/anteater2/bitmesh/rpc"
	"github.com/anteater2/chord-node/config"
	"github.com/anteater2/chord-node/utils"
)

var Address string
var Key key.Key
var Fingers = make([]*RemoteNode, config.NumFingers())
var Predecessor *RemoteNode
var Successor *RemoteNode
var RPCCaller *rpc.Caller
var RPCCallee *rpc.Callee

var RPCFindSuccessor rpc.RemoteFunc
var RPCNotify rpc.RemoteFunc
var RPCGetPredecessor rpc.RemoteFunc

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Address string
	Key     key.Key
}

// ClosestPrecedingNode finds the closest preceding node to the key in this node's finger table.
// This doesn't need any RPC.
func ClosestPrecedingNode(key key.Key) RemoteNode {
	for i := config.NumFingers() - 1; i >= 0; i-- {
		if Fingers[i] == nil {
			panic("You attempted to find ClosestPrecedingNode without an initialized finger table!")
		}
		if key.BetweenExclusive(Key, key) {
			return *Fingers[i]
		}
	}
	return RemoteNode{Address: Address, Key: Key}
}

// FindSuccessor finds the successor node to the key.  This may require RPC calls.
func FindSuccessor(key key.Key) RemoteNode {
	if key.BetweenEndInclusive(Key, Successor.Key) {
		// key is between this node and its successor
		return RemoteNode{Address: Address, Key: Key}
	}
	target := ClosestPrecedingNode(key)
	// Now, we have to do an RPC on target to find the successor.
	interf, err := RPCFindSuccessor(target.Address+":2000", uint64(key))
	if err != nil {
		panic("RPC FindSuccessor Failed!")
	}
	rv := interf.(RemoteNode)
	return rv
}

// Notify notifies the successor that you are the predecessor
func Notify(node RemoteNode) int {
	if Predecessor == nil {
		Predecessor = &node
	}
	return 0 // Necessary to interface with RPCCaller
}

func GetPredecessor() RemoteNode {
	return *Predecessor
}

//Stabilize the Successor and Predecessor fields of this node.
//This is a goroutine and never terminates.
func Stabilize() {
	Successor.Address
}

// CreateLocalNode creates a local node on its own ring.  It can be inserted into another ring later.
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

	Address = GetOutboundIP()
	Predecessor = nil
	Successor = &RemoteNode{
		Address: Address,
		Key:     Key,
	}
	// Initialize the finger table for the solo ring configuration
	for i := uint64(0); i < config.NumFingers(); i++ {
		Fingers[i] = Successor
	}

	// Define all of the RPC functions.  For more info, look at Yuchen's caller.go and example_test.go
	RPCFindSuccessor = RPCCaller.Declare(key.NewKey(1), RemoteNode{}, 10)
	RPCNotify = RPCCaller.Declare(RemoteNode{}, 0, 10)
	RPCGetPredecessor = RPCCaller.Declare(interface{}, RemoteNode{}, 10)
	// RPCIsAlive = RPCCaller.Declare(nil{}, nil{}, 0) // Don't define this just yet - how does the RPC system react if a node fails to respond?

	// Hook the RPCCallee into this node's functions
	RPCCallee.Implement(FindSuccessor)
	RPCCallee.Implement(Notify)
}

// Join a ring given a node IP address.
func Join(ring string) {
	ringCallee := ring + string(config.CalleePort())
	ringSuccessorInterf, err := RPCFindSuccessor(ringCallee, Key)
	if err != nil {
		panic("RPCFindSuccessor failed!")
	}
	ringSuccessor := ringSuccessorInterf.(RemoteNode)
	Successor = &ringSuccessor
}

// GetOutboundIP gets preferred outbound IP of this machine using a filthy hack
// The connection should not actually require the Google DNS service (the 8.8.8.8),
// but by creating it we can see what our preferred IP is.
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
func main() {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := utils.LocalAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		fmt.Printf("%s\n", addr.String())
	}

	fmt.Printf("Creating local node on its own ring...\n")
	CreateLocalNode()
	go RPCCallee.Start()
	go RPCCaller.Start()
	if !config.Creator() {
		fmt.Printf("Connecting node to network at %s\n", config.Introducer())
		Join(config.Introducer())
	}

}
