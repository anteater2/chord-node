// LinkedList Bucketed Hash Table
// Borrowed from https://gist.github.com/urielhdz/25a86726bce759444255
package table

import (
	"errors"

	"../key"
)

type HashEntry struct {
	value []byte
	key   string
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

func (self *HashTable) Put(hashKey string, value []byte) {
	// TO DO: Replace if key is the same
	position := key.Hash(hashKey, self.maximum)
	newHashEntry := HashEntry{key: hashKey, value: value}
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
		if hashEntry.key == hashKey {
			return hashEntry.value, nil
		}
		if hashEntry.next == nil {
			break
		}
		hashEntry = *hashEntry.next
	}
	return []byte{0}, errors.New("No such key!")
}
func (self HashEntry) IsNil() bool {
	return len(self.value) == 0 && self.key == ""
}
