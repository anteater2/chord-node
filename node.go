package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"strconv"
	"time"

	"./config"
	"./key"
	"./src/github.com/anteater2/bitmesh/rpc"
)

var Address string
var Key key.Key
var Fingers []*RemoteNode
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
	for i := config.NumFingers() - 1; i > 0; i-- { // WARNING: GO DOES THIS i>0 CHECK AT THE END OF THE LOOP!
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
	interf, err := RPCFindSuccessor(target.Address+":"+strconv.Itoa(config.CalleePort()), key)
	if err != nil {
		log.Printf("[DIAGNOSTIC] Remote target is " + target.Address + ":" + strconv.Itoa(config.CalleePort()) + "\n")
		log.Print(err)
		panic("RPCFindSuccessor failed!")
	}
	rv := interf.(RemoteNode)
	return rv
}

// Notify notifies the successor that you are the predecessor
func Notify(node RemoteNode) int {
	fmt.Printf("Got notify from %s!\n", node.Address)
	if Predecessor == nil || node.Key.BetweenExclusive(Key, Successor.Key) {
		Predecessor = &node
	}
	return 0 // Necessary to interface with RPCCaller
}

//Stabilize the Successor and Predecessor fields of this node.
//This is a goroutine and never terminates.
func Stabilize() {
	for true { // This is how while loops work.  Not even joking.
		if Successor.Address != Address {
			fmt.Printf("Stabilizing!\n")
			remoteInterf, err := RPCGetPredecessor(Successor.Address+":"+strconv.Itoa(config.CalleePort()), 0) // 0 is a dummy value so that the RPC interface can work
			if err != nil {
				log.Printf("[DIAGNOSTIC] Stabilization call failed!")
				log.Print(err)
				panic("RPCGetPredecessor failed!")
			}
			remote := remoteInterf.(RemoteNode)
			fmt.Printf("Remote key is %d\n", remote.Key)
			fmt.Printf("My keyspace is (%d, %d)\n", Key, Successor.Key)
			if remote.Key.BetweenExclusive(Key, Successor.Key) {
				Successor = &remote
			}

			RPCNotify(Successor.Address+":"+strconv.Itoa(config.CalleePort()), RemoteNode{
				Address: Address,
				Key:     Key,
			})
		} else {
			//fmt.Printf("Skipping stabilization, ring is serviced only by one node.\n")
		}
		time.Sleep(time.Second * 1)
	}
}

//FixFingers is the finger-table updater.
//Again, this is a goroutine and never terminates.
func FixFingers() {
	fmt.Printf("Starting to finger nodes...\n") //hehehe
	currentFingerIndex := uint64(0)
	for true { //Again, a while loop in the "simple" and "easy to read" language
		currentFingerIndex++
		currentFingerIndex %= config.NumFingers()
		//fmt.Printf("Updating finger %d of %d\n", currentFingerIndex, len(Fingers))
		offset := uint64(math.Pow(2, float64(currentFingerIndex)))
		val := (uint64(Key) + offset) % config.NumFingers()
		newFinger := FindSuccessor(key.NewKey(val))
		Fingers[currentFingerIndex] = &newFinger
		time.Sleep(time.Second * 1)
	}
}

// CreateLocalNode creates a local node on its own ring.  It can be inserted into another ring later.
func CreateLocalNode() {
	// Set the variables of this node.
	var err error
	RPCCaller, err = rpc.NewCaller(config.CallerPort())
	if err != nil {
		panic("RPCCaller failed to initialize")
	}
	RPCCallee, err = rpc.NewCallee(config.CalleePort())
	if err != nil {
		panic("RPCCallee failed to initialize")
	}

	Address = config.Addr()

	Key = key.Key(hash(Address) % config.MaxKey())
	fmt.Printf("Keyspace position %d was derived from IP%s\n", Key, config.Addr())

	Predecessor = nil
	Successor = &RemoteNode{
		Address: Address,
		Key:     Key,
	}
	// Initialize the finger table for the solo ring configuration
	Fingers = make([]*RemoteNode, config.NumFingers())
	fmt.Printf("Finger table size %d was derived from the keyspace size\n", config.NumFingers())
	for i := uint64(0); i < config.NumFingers(); i++ {
		Fingers[i] = Successor
	}

	// Define all of the RPC functions.
	// For more info, look at Yuchen's caller.go and example_test.go
	// Go's type "system" is going to make me kill myself.
	RPCNotify = RPCCaller.Declare(RemoteNode{}, 0, 10*time.Second)
	RPCFindSuccessor = RPCCaller.Declare(key.NewKey(1), RemoteNode{}, 10*time.Second)
	RPCGetPredecessor = RPCCaller.Declare(0, RemoteNode{}, 10*time.Second)
	// RPCIsAlive = RPCCaller.Declare(nil{}, nil{}, 0) // Don't define this just yet - how does the RPC system react if a node fails to respond?

	// Hook the RPCCallee into this node's functions
	RPCCallee.Implement(FindSuccessor)
	RPCCallee.Implement(Notify)
	RPCCallee.Implement(GetPredecessor)
}

//GetPredecessor is a getter for the predecessor, implemented for the sake of RPC calls.
//Note that the RPC calling interface does not allow argument-free functions, so this takes
// a worthless int as argument.
func GetPredecessor(void int) RemoteNode {
	fmt.Printf("RPC Call to GetPredecessor!\n")
	if Predecessor == nil {
		fmt.Printf("Returned self node, no predecessor set.\n")
		return RemoteNode{
			Address: Address,
			Key:     Key,
		}
	}
	fmt.Printf("Returned predecessor.\n")
	return *Predecessor
}

// Join a ring given a node IP address.
func Join(ring string) {
	ringCallee := ring + ":" + strconv.Itoa(config.CalleePort())
	ringSuccessorInterf, err := RPCFindSuccessor(ringCallee, Key)
	if err != nil {
		log.Printf("[DIAGNOSTIC] Join failed.  Target: %s", ringCallee)
		log.Print(err)
		panic("RPCFindSuccessor failed!")
	}
	ringSuccessor := ringSuccessorInterf.(RemoteNode)
	Successor = &ringSuccessor
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Creating local node @IP%s on its own ring of size %d...\n", config.Addr(), config.MaxKey())
	CreateLocalNode()
	go RPCCallee.Start()
	go RPCCaller.Start()
	if !config.Creator() {
		fmt.Printf("Connecting node to network at %s\n", config.Introducer())
		Join(config.Introducer())
	}
	fmt.Printf("Beginning stabilizer...\n")
	go Stabilize()
	go FixFingers()
	select {} // Wait forever so that the goroutines never terminate
}
