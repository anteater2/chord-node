package config

import "github.com/anteater2/chord-node/utils"

const bits uint64 = 10

// CallerPort is the port of the caller
const CallerPort = 2000

// CalleePort is the port of the callee
const CalleePort = 2001

// MaxKey is the size of the key space. Can't be constant since its the result
// of a function call
var MaxKey = utils.IntPow(2, 10)

// NumFingers the size of any individual finger table
var NumFingers = bits - 1
