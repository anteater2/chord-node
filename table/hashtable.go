// LinkedList Bucketed Hash Table
// Borrowed from https://gist.github.com/urielhdz/25a86726bce759444255
package table

import (
	"errors"

	"github.com/anteater2/chord-node/key"
)

type HashEntry struct {
	Value []byte
	Key   string
	next  *HashEntry
}

// HashTable is a hash table mapping strings to byte arrays.
// lol no generics
type HashTable struct {
	hashEntries []HashEntry
	maximum     uint64
}

func NewTable(maxKeys uint64) *HashTable {
	return &HashTable{maximum: maxKeys, hashEntries: make([]HashEntry, maxKeys)}
}

func (self *HashTable) GetRange(start key.Key, end key.Key) []HashEntry {
	entries := []HashEntry{}
	for i := start + 1; i <= end+1; i++ {
		hashEntry := &self.hashEntries[i]
		if !hashEntry.IsNil() {
			entries = append(entries, *hashEntry)
			for hashEntry.next != nil {
				hashEntry = hashEntry.next
				entries = append(entries, *hashEntry)
			}
		}
	}
	return entries
}

func (self *HashTable) Put(hashKey string, value []byte) {
	// TO DO: Replace if key is the same
	position := key.Hash(hashKey, self.maximum)
	newHashEntry := HashEntry{Key: hashKey, Value: value}
	hashEntry := &self.hashEntries[position]
	if hashEntry.IsNil() {
		self.hashEntries[position] = newHashEntry
	} else {
		for hashEntry.next != nil {
			hashEntry = hashEntry.next
		}
		hashEntry.next = &newHashEntry
	}
}
func (self *HashTable) Get(hashKey string) ([]byte, error) {
	position := key.Hash(hashKey, self.maximum)
	hashEntry := self.hashEntries[position]
	for !hashEntry.IsNil() {
		if hashEntry.Key == hashKey {
			return hashEntry.Value, nil
		}
		if hashEntry.next == nil {
			break
		}
		hashEntry = *hashEntry.next
	}
	return []byte{0}, errors.New("No such key!")
}
func (self HashEntry) IsNil() bool {
	return self.Value == nil && self.Key == ""
}
