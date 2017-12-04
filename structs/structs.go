package structs

import "github.com/anteater2/chord-node/key"

// RemoteNode holds information for connecting to a remote node
type RemoteNode struct {
	Address string
	Key     key.Key
}

type GetKeyResponse struct {
	Data  []byte
	Error bool
}

type PutKeyRequest struct {
	KeyString string
	Data      []byte
}

type GetKeyRangeRequest struct {
	Start key.Key
	End   key.Key
}
