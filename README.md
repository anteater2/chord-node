# Bitmesh Chord Node
## Structure
```node.go``` is the only file that is responsible for everything.
## Protocol
Do RPC using Yuchen's library.  Implement the functions in the Chord paper:

https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf

The idea is that each node should implement callee receivers for:

```FindSuccessor(key uint32)```

```Notify(node RemoteNode)```

```IsAlive() //but this one might be hard```

They should also define caller interfaces for each. The periodic functions:

```Stabilize```

```FixFingers```

```CheckPredecessor // again, this is hard```

can be goroutines with sleeps.

## IP Resolution
Each node needs to know its own IP, because the RPC library definitely doesn't.

## Ports
Callers send on port 2000.

Callees recv on port 2001.

## Other problems
I set the number of fingers to be exactly 3 because I don't know how to get the size of the ring.  I don't think that this should cause problems, but it might break something.

The node program isn't done, but it shouldn't be too bad (especially because we have 5ish days to finish it)

The node program needs to take some command-line options like the ring to join too.

Hopefully, this is enough to work on for the next few days.