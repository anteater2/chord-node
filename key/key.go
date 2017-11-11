package key

import "github.com/anteater2/chord-node/config"

// Key is a key in the distributed hash table
type Key uint64

// Between returns true if key is between keys start and end i.e. within
// (s, e), false otherwise
func (key Key) Between(start Key, end Key) bool {
	s, e := uint64(start), uint64(end)
	return key.InBounds(s+1, e)
}

// InBounds returns true if key is within [start, end), false otherwise
// If start > end, then the interval wraps around
// e.g. [start ... MaxKey - 1, 0, 1 ... end])
func (key Key) InBounds(start uint64, end uint64) bool {
	k := uint64(key)

	if start > config.MaxKey() || end > config.MaxKey() {
		panic("MaxKey constraint violated by start")
	}

	if start == end {
		panic("Invalid arguments")
	} else if start > end { // Interval wraps
		return start <= k || k <= end
	} else {
		return start <= uint64(key) && uint64(key) < end
	}
}

// NewKey returns a new key
func NewKey(value uint64) Key {
	return Key(value)
}

// Valid returns true if the key is within the keyspace, false otherwise
func (key Key) Valid() bool {
	k := uint64(key)
	return 0 <= k && k < config.MaxKey() // Check for 0 <= not necessary
}
