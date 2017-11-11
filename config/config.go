package config

import (
	"errors"
	"flag"
	"net"

	"github.com/anteater2/chord-node/utils"
)

var (
	addr       net.Addr
	bits       uint64
	callerPort = 2000
	calleePort = 2001
	introducer string
	isCreator  bool
	maxKey     uint64
	numFingers uint64
	username   string
)

// Addr returns the local address
func Addr() net.Addr {
	return addr
}

// CallerPort returns the port of the caller
func CallerPort() int {
	return callerPort
}

// CalleePort returns the port of the callee
func CalleePort() int {
	return calleePort
}

// Creator returns the truth value of whether this node is the first node
// in a chord network
func Creator() bool {
	return isCreator
}

// Init initializes the configs
func Init() error {
	addrs, err := utils.LocalAddrs()
	if err != nil {
		return err
	}

	addr = addrs[0] // TODO: better logic for which of addrs to use

	flag.Uint64Var(
		&bits,
		"n",
		0,
		"create a new chord network with a keyspace of 2**numBits",
	)

	flag.StringVar(
		&introducer,
		"c",
		"",
		"create a new node and connect to the specified address",
	)

	flag.StringVar(
		&username,
		"u",
		"",
		"the username to use",
	)

	flag.Parse()

	if bits == 0 && introducer == "" {
		return errors.New("Need to either create or connect")
	}

	if bits != 0 && introducer != "" {
		return errors.New("Cannot create and connect at same time")
	}

	if username == "" {
		return errors.New("No username")
	}

	if bits != 0 && introducer == "" {
		if bits > 63 {
			return errors.New("Maximum bits: 63") // Not really, but easier for now
		}
		isCreator = true
		maxKey = utils.IntPow(2, bits)
		numFingers = bits - 1
		return nil
	}

	isCreator = false
	return nil
}

// Introducer returns the introducing address
func Introducer() string {
	return introducer
}

// MaxKey returns the size of the key space.
func MaxKey() uint64 {
	return maxKey
}

// NumFingers returns the size of a finger table
func NumFingers() uint64 {
	return numFingers
}

// Username returns the username of the node
func Username() string {
	return username
}
