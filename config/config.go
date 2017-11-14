package config

import (
	"errors"
	"flag"
	"log"
	"net"

	"../utils"
)

var (
	addr       string
	bits       uint64
	callerPort = 2000
	calleePort = 2001
	introducer string
	isCreator  bool
	maxKey     uint64
	numFingers uint64
	username   string
)

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

// Addr is this node's preferred outbound IP
func Addr() string {
	return addr
}

// Init initializes the configs
func Init() error {
	/*addrs, err := utils.LocalAddrs()
	if err != nil {
		return err
	}
	*/
	addr = GetOutboundIP()
	flag.Uint64Var(
		&bits,
		"n",
		0,
		"Create a new chord ring with a keyspace of size 2^numBits",
	)

	flag.StringVar(
		&introducer,
		"c",
		"",
		"Create a new node and connect to the specified ring address",
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

	/*if bits != 0 && introducer != "" {
		return errors.New("Cannot create and connect at same time")
	}*/

	/*if username == "" {
		return errors.New("No username")
	}*/
	if bits == 0 {
		return errors.New("You must specify a valid number of bits!")
	}

	if bits > 63 {
		return errors.New("Maximum bits: 63") // Not really, but easier for now
	}
	isCreator = introducer == ""
	maxKey = utils.IntPow(2, bits)
	numFingers = bits - 1
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
