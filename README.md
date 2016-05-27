TRR: A Key-Value time-series with gorilla algorithm using in Raft consistency RPC Server
==============

[![GoDoc](https://godoc.org/github.com/kkdai/trr?status.svg)](https://godoc.org/github.com/kkdai/trr)  [![Build Status](https://travis-ci.org/kkdai/trr.svg?branch=master)](https://travis-ci.org/kkdai/trr)



What is TRR
=============


TRR (Time-series Raft RPC client/server) is a package to help you hosted a simple KV value with time-series data under [raft consensus algorithm](https://github.com/coreos/etcd). (implement by [CoreOS/etcd](https://github.com/coreos/etcd)).

It provide a basic RPC Client/Server for K/V(Key Value) storage service.

Features
=============

- [raft consensus algorithm](https://github.com/coreos/etcd)
- Key/Value base usage, easy to Get/Set time-series data.
- Based on Gorilla algorithm which could reduce data size to 12X.
- RPC entry point, easy to use.


What is Raft
=============

Raft is a consensus algorithm that is designed to be easy to understand. It's equivalent to Paxos in fault-tolerance and performance. The difference is that it's decomposed into relatively independent subproblems, and it cleanly addresses all major pieces needed for practical systems. We hope Raft will make consensus available to a wider audience, and that this wider audience will be able to develop a variety of higher quality consensus-based systems than are available today. (quote from [here](https://raft.github.io/))

How to use etcd/raft in your project
=============

1. Refer code from [raftexample](https://github.com/coreos/etcd/tree/master/contrib/raftexample)
2. Get file listener.go, kvstore.go, raft.go. 
3. Do your modification for your usage.

### note
- `raft.transport` need an extra http port for raft message exchange. **MUST** add this in your code. (which is peer info in example code)


Installation and Usage
=============


Install
---------------
```
go get github.com/kkdai/trr
```

Usage
---------------

### Server Example(1) Single Server:

```go
package main
    
import (
	"fmt"
    
	. "github.com/kkdai/trr"
)
    
func main() {
	forever := make(chan int)

	//RPC addr
	rpcAddr := "127.0.0.1:1234"
	srv := StartServer(rpcAddr, 1)

	<-forever
}
```

### Server Example(2) Cluster Server:

```go
package main
    
import (
	"fmt"
    
	. "github.com/kkdai/raftrpc"
)
    
func main() {
	forever := make(chan int)
	
	//Note there are two address and port.
	//
	// "127.0.0.1:1234" is RPC access point
	// "http://127.0.0.1:12379" is raft message access point which use http
	
	var raftMsgSrvList []string
	raftMsgSrvList = append(raftMsgSrvList, "http://127.0.0.1:12379")
	raftMsgSrvList = append(raftMsgSrvList, "http://127.0.0.1:22379")

	srv1 := StartClusterServers("127.0.0.1:1234", 1, raftMsgSrvList)
	srv2 := StartClusterServers("127.0.0.1:1235", 2, raftMsgSrvList)
	
	<-forever
}
```

### Client Example

Assume a server exist on `127.0.0.1:1234`.


```go
package main
    
import (
	"fmt"
    "log"
    
	. "github.com/kkdai/raftrpc"
)
    
func main() {
	client := MakeClerk("127.0.0.1:1234")
	t0, _ := time.ParseInLocation("Jan _2 2006 15:04:05", "Mar 24 2015 02:00:00", time.Local)
	t0unix := uint32(t0.Unix())

	srv := StartServer("127.0.0.1:1230", 1)

	client := MakeClerk("127.0.0.1:1230")
	client.PutTimeData("t1", t0unix, 10)

	t1unix := t0unix + 62
	client.PutTimeData("t1", t1unix, 12)

	t2unix := t1unix + 62
	client.PutTimeDataBack("t1", t2unix, 14)
	
	tt, vv, err := client.GetTimeData("t1")
	if err != nil || tt != t0unix || vv != 10 {
		log.Println("Simple time get error", tt, vv, err)
	}
}	
```

Inspired By
---------------
- [CoreOS ETCD source code](https://github.com/coreos/etcd)
- [ETCD Example](https://github.com/coreos/etcd/tree/master/contrib/raftexample)
- [Raft: A First Implementation](http://otm.github.io/2015/05/raft-a-first-implementation/)
- [Gorilla time-series algorithm on golang](https://github.com/dgryski/go-tsz)

Project52
---------------

It is one of my [project 52](https://github.com/kkdai/project52).


License
---------------

etcd is under the Apache 2.0 [license](LICENSE). See the LICENSE file for details.