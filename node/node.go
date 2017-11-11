package main

import "github.com/anteater2/bitmesh/rpc"

const (
	MaxKey = 1024
)

// LocalNode holds the information required to run a node
type LocalNode struct {
	Address     string
	Key         uint32
	Fingers     []RemoteNode
	Predecessor RemoteNode
	Successor   RemoteNode
	RPCCaller   rpc.Caller
	RPCCallee   rpc.Callee
}

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Address string
	Key     uint32
}

type KeyArg struct {
	Key uint32
}

func isInEndInclusive(key uint32, start uint32, end uint32) bool {
	key = key % MaxKey
	if start > MaxKey {
		panic("MaxKey constraint violated by start")
	}
	if end > MaxKey {
		panic("MaxKey constraint violated by end")
	}
	if start < end {
		return key > start && key <= end
	}
	if start >= end {
		return key > start || key <= end
	}
	return false // Jesus what a stupid compiler
}

func isInExclusive(key uint32, start uint32, end uint32) bool {
	key = key % MaxKey
	if start > MaxKey {
		panic("MaxKey constraint violated by start")
	}
	if end > MaxKey {
		panic("MaxKey constraint violated by end")
	}
	if start < end {
		return key > start && key < end
	}
	if start >= end {
		return key > start || key < end
	}
	return false // Jesus what a stupid compiler
}

func (node *LocalNode) findSuccessor(key uint32) RemoteNode {
	if isInEndInclusive(key, node.Key, node.Successor.Key) {
		return RemoteNode{node.Address, node.Key}
	} else {
		// node.RPCCaller.Declare(KeyArg{}, RemoteNode{}, time.sec)
	}
}
