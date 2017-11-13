# Bitmesh Chord Node
## Structure
```node.go``` is the only file that is responsible for everything.
## Protocol
RPC uses Yuchen's library.
All RPC calls (for now) are documented in the MIT Chord Paper:

https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf

## Implementation
Each node implements callee receivers for:

```FindSuccessor(key uint32)```

```Notify(node RemoteNode)```

```GetPredecessor()```

```IsAlive() //but this one might be hard```

They should also define caller interfaces for each. The periodic functions:

```Stabilize```

```FixFingers```

```CheckPredecessor // again, this is hard, and currently not implemented```

are be goroutines with sleeps.

## IP Resolution
Each node needs to know its own IP, because the RPC library definitely doesn't.
This is handled in a really hacky and sort of disgusting way: we attempt to connect to Google's DNS servers at 8.8.8.8.
We don't actually care about the DNS server, but when the connection is made we can snoop to see what local IP is bound to it.

## Ports
Callers send on port 2000.
Callees receive on port 2001.

# Docker Testing
## Building the image
The Dockerfile will build a docker image using Go 1.9.  It does *not* go to Yuchen's repo to get the RPC library (those Github pulls give me anxiety and also break half the time, so all imports here are local) - instead, it needs the library to be in src/github.com/... 
Builds can be done with:

```docker build -t bitmesh-node .```

At some point, it might make sense to push this to a docker repo to make it easier, but for now this'll do.

## Networking
Docker defines a bridge0 interface to "bridge" the container VMs with the host OS. This bridge0 interface connects on a virtual bridge network that, by default, is shared amongst all the containers running on a system.  So, if we start two docker containers, they will both run on the same network, with different virtual IP addresses.

The current version of this has the RPC call to GetPredecessor fail (the call is made from node.go:81), which is really weird because the callee always seems to recognize the call and return a value (there are print statements that fire). It's 2AM now, so I'm just committing this and being done.

It should definitely be possible to finish this by Wednesday, though.

### Core Node

You can start a core node using the Ubuntu version of docker with:

```docker run -it -p2000:2000 -p2001:2001 bitmesh-node -n 10```

The ```-pXXXX:YYYY``` flags map the container ports to the host's ports - you should be able to connect to this from even your local machine, but IP subnetting issues will probably make this a bad idea.

### Adding extra nodes
All nodes will print their IP when starting.
You can connect a new node to an existing one using:

```docker run -it bitmesh-node -n 10 -c [IP HERE]```

Hopefully there is enough 

# Other problems
- We really need to start thinking about key handoff
- We also need to start considering reliability