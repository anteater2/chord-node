# Bitmesh Chord Node
## Structure
```main.go``` handles the actual responsibilities of setting up a server socket and handling RPC calls.
```node.go``` contains structs and functions for mutating node state (successor, finger table, etc.)
## Protocol
Do RPC using Yuchen's library.  Implement the functions in the Chord paper:
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf