package config

import "github.com/anteater2/chord-node/utils"

var (
	bits       uint64 = 10
	callerPort        = 2000
	calleePort        = 2001
	maxKey            = utils.IntPow(2, bits)
	numFingers        = bits - 1
)

// CallerPort returns the port of the caller
func CallerPort() int {
	return callerPort
}

// CalleePort returns the port of the callee
func CalleePort() int {
	return calleePort
}

// MaxKey returns the size of the key space.
func MaxKey() uint64 {
	return maxKey
}

// NumFingers returns the size of a finger table
func NumFingers() uint64 {
	return numFingers
}
