package main

import (
	"hash/fnv"
	"log"
	"math"
	"strconv"
	"time"

	"./config"
	"./key"
	"github.com/anteater2/bitmesh/rpc"
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
var RPCIsAlive rpc.RemoteFunc

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Address string
	Key     key.Key
}

// ClosestPrecedingNode finds the closest preceding node to the key in this node's finger table.
// This doesn't need any RPC.
func ClosestPrecedingNode(key key.Key) RemoteNode {
	for i := config.NumFingers() - 1; i > 0; i-- { // WARNING: GO DOES THIS i>0 CHECK AT THE END OF THE LOOP!
		//log.Printf("Checking finger %d\n", i)
		if Fingers[i] == nil {
			panic("You attempted to find ClosestPrecedingNode without an initialized finger table!")
		}
		if Fingers[i].Key.BetweenExclusive(Key, key) {
			return *Fingers[i]
		}
	}
	return RemoteNode{Address: Address, Key: Key}
}

// FindSuccessor finds the successor node to the key.  This may require RPC calls.
func FindSuccessor(key key.Key) RemoteNode {
	if key.BetweenEndInclusive(Key, Successor.Key) {
		// key is between this node and its successor
		return *Successor
	}
	target := ClosestPrecedingNode(key)
	if target.Address == Address {
		log.Printf("[DIAGNOSTIC] Infinite loop detected!\n")
		log.Printf("[DIAGNOSTIC] This is likely because of a bad finger table.\n")
		panic("This is probably a serious bug.")
	}
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
	if Predecessor == nil || node.Key.BetweenExclusive(Predecessor.Key, Key) {
		log.Printf("Got notify from %s!  New predecessor: %d\n", node.Address, node.Key)
		Predecessor = &node
	}
	return 0 // Necessary to interface with RPCCaller
}

//Stabilize the Successor and Predecessor fields of this node.
//This is a goroutine and never terminates.
func Stabilize() {
	for true { // This is how while loops work.  Not even joking.
		var remote RemoteNode
		if Predecessor == nil {
			log.Printf("Null predecessor!  New predecessor: %d\n", Successor.Key)
			Predecessor = Successor
		}
		if Successor.Address == Address {
			// Avoid making an RPC call to ourselves
			remote = *Predecessor
		} else {
			remoteInterf, err := RPCGetPredecessor(Successor.Address+":"+strconv.Itoa(config.CalleePort()), 0) // 0 is a dummy value so that the RPC interface can work
			if err != nil {
				log.Printf("[DIAGNOSTIC] Stabilization call failed!")
				remote := remoteInterf.(RemoteNode)
				log.Printf("[DIAGNOSTIC] Returned result: " + strconv.Itoa(int(remote.Key)))
				log.Print(err)
				panic("RPCGetPredecessor failed!")
			}
			remote = remoteInterf.(RemoteNode)
		}
		if remote.Key.BetweenExclusive(Key, Successor.Key) {
			log.Printf("My keyspace is (%d, %d)\n", Key, Successor.Key)
			log.Printf("New successor %d\n", remote.Key)
			Successor = &remote
			Fingers[0] = &remote
		}

		RPCNotify(Successor.Address+":"+strconv.Itoa(config.CalleePort()), RemoteNode{
			Address: Address,
			Key:     Key,
		})
		time.Sleep(time.Second * 1)
	}
}

//FixFingers is the finger-table updater.
//Again, this is a goroutine and never terminates.
func FixFingers() {
	log.Printf("Starting to finger nodes...\n") //hehehe
	currentFingerIndex := uint64(0)
	for true {
		currentFingerIndex++
		currentFingerIndex %= config.NumFingers()
		offset := uint64(math.Pow(2, float64(currentFingerIndex)))
		val := (uint64(Key) + offset) % config.MaxKey()
		newFinger := FindSuccessor(key.NewKey(val))
		//log.Printf("Updating finger %d (pointing to key %d) of %d to point to node %s\n", currentFingerIndex, val, len(Fingers), newFinger.Address)
		if newFinger.Address != Fingers[currentFingerIndex].Address {
			log.Printf("Updating finger %d (pointing to key %d) of %d to point to node %s\n", currentFingerIndex, val, len(Fingers)-1, newFinger.Address)
		}
		Fingers[currentFingerIndex] = &newFinger
		time.Sleep(time.Second * 1)
	}
}

// IsAlive is a heartbeat check.  If this fails, the RPC call will err out.
func IsAlive(void bool) bool {
	return void
}

// CheckPredecessor is a goroutine that keeps tabs on the predecessor and updates itself if the predecessor leaves the network.
// Currently not working - RPCIsAlive is not timing out!
func CheckPredecessor() {
	for true {
		if Predecessor != nil {
			resp, err := RPCIsAlive(Predecessor.Address+":"+strconv.Itoa(config.CalleePort()), true)
			if err != nil {
				log.Printf("Predecessor " + Predecessor.Address + " failed a health check!  Attempting to adjust...")
				log.Print(err)
				Predecessor = nil
			}
		}
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
	log.Printf("Keyspace position %d was derived from IP%s\n", Key, config.Addr())

	Predecessor = nil
	Successor = &RemoteNode{
		Address: Address,
		Key:     Key,
	}
	// Initialize the finger table for the solo ring configuration
	Fingers = make([]*RemoteNode, config.NumFingers())
	log.Printf("Finger table size %d was derived from the keyspace size\n", config.NumFingers())
	for i := uint64(0); i < config.NumFingers(); i++ {
		Fingers[i] = Successor
	}

	// Define all of the RPC functions.
	// For more info, look at Yuchen's caller.go and example_test.go
	// Go's type "system" is going to make me kill myself.
	RPCNotify = RPCCaller.Declare(RemoteNode{}, 0, 10*time.Second)
	RPCFindSuccessor = RPCCaller.Declare(key.NewKey(1), RemoteNode{}, 10*time.Second)
	RPCGetPredecessor = RPCCaller.Declare(0, RemoteNode{}, 10*time.Second)
	RPCIsAlive = RPCCaller.Declare(true, true, 1*time.Second)

	// Hook the RPCCallee into this node's functions
	RPCCallee.Implement(FindSuccessor)
	RPCCallee.Implement(Notify)
	RPCCallee.Implement(GetPredecessor)
	RPCCallee.Implement(IsAlive)
}

//GetPredecessor is a getter for the predecessor, implemented for the sake of RPC calls.
//Note that the RPC calling interface does not allow argument-free functions, so this takes
//a worthless int as argument.
func GetPredecessor(void int) RemoteNode {
	//log.Printf("RPC Call to GetPredecessor!\n")
	if Predecessor == nil {
		//log.Printf("Returned self node, no predecessor set.\n")
		return RemoteNode{
			Address: Address,
			Key:     Key,
		}
	}
	//log.Printf("Returned predecessor.\n")
	return *Predecessor
}

// Join a ring given a node IP address.
func Join(ring string) {
	log.Printf("Connecting node to network at %s\n", config.Introducer())
	ringCallee := ring + ":" + strconv.Itoa(config.CalleePort())
	ringSuccessorInterf, err := RPCFindSuccessor(ringCallee, Key)
	if err != nil {
		log.Printf("[DIAGNOSTIC] Join failed.  Target: %s", ringCallee)
		log.Print(err)
		panic("RPCFindSuccessor failed!")
	}
	ringSuccessor := ringSuccessorInterf.(RemoteNode)
	Successor = &ringSuccessor
	Fingers[0] = &ringSuccessor
	log.Printf("New successor %d!\n", Successor.Key)
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
	log.Printf("Creating local node @IP%s on its own ring of size %d...\n", config.Addr(), config.MaxKey())
	CreateLocalNode()
	go RPCCallee.Start()
	go RPCCaller.Start()
	if !config.Creator() {
		Join(config.Introducer())
	}
	log.Printf("Beginning stabilizer...\n")
	go Stabilize()
	go FixFingers()
	go CheckPredecessor()
	select {} // Wait forever so that the goroutines never terminate
}
