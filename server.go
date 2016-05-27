// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trr

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"syscall"

	"github.com/coreos/etcd/raft/raftpb"
)

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type KVRaft struct {
	mu         sync.Mutex
	l          net.Listener
	me         int
	dead       bool // for testing
	unreliable bool // for testing

	// Your definitions here.
	store    *kvstore
	srvStopc chan struct{}
}

func (kv *KVRaft) Get(args *GetArgs, reply *GetReply) error {
	log.Println("[GET]", args)

	if args.Key == "" {
		log.Println("[GET]", InvalidParam)
		return errors.New(InvalidParam)
	}

	if v, ok := kv.store.Lookup(args.Key); ok {
		reply.Value = v
		return nil
	}

	reply.Err = ErrNoKey
	return errors.New(ErrNoKey)
}

func (kv *KVRaft) Put(args *PutArgs, reply *PutReply) error {
	log.Println("[PUT]", args)

	if args.Key == "" || args.Value == "" {
		log.Println("[PUT]", InvalidParam)
		err := errors.New(InvalidParam)
		reply.Err = InvalidParam
		return err
	}

	if v, ok := kv.store.Lookup(args.Key); ok {
		reply.PreviousValue = v
	}

	reply.Err = "NIL"
	log.Println("[PUT] ", args)
	kv.store.Propose(args.Key, args.Value)
	return nil
}

// tell the server to shut itself down.
// please do not change this function.
func (kv *KVRaft) kill() {
	DPrintf("Kill(%d): die\n", kv.me)
	kv.dead = true
	close(kv.srvStopc)
	//remove socket file
}

func StartServer(rpcPort string, me int) *KVRaft {
	return startServer(rpcPort, me, []string{rpcPort}, false)
}

func StartClusterServers(rpcPort string, me int, cluster []string) *KVRaft {
	return startServer(rpcPort, me, cluster, false)
}

func StarServerJoinCluster(rpcPort string, me int) *KVRaft {
	return startServer(rpcPort, me, []string{rpcPort}, true)
}

func startServer(serversPort string, me int, cluster []string, join bool) *KVRaft {
	//gob.Register(Op{})

	kv := new(KVRaft)
	rpcs := rpc.NewServer()
	rpcs.Register(kv)

	proposeC := make(chan string)
	//defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	//defer close(confChangeC)

	//node
	commitC, errorC, stopc := newRaftNode(me, cluster, join, proposeC, confChangeC)
	kv.srvStopc = stopc

	//kvstore
	kv.store = newKVStore(proposeC, commitC, errorC)

	log.Println("[server] ", me, " ==> ", serversPort)
	l, e := net.Listen("tcp", serversPort)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	kv.l = l

	go func() {
		for kv.dead == false {
			conn, err := kv.l.Accept()
			if err == nil && kv.dead == false {
				if kv.unreliable && (rand.Int63()%1000) < 100 {
					// discard the request.
					conn.Close()
				} else if kv.unreliable && (rand.Int63()%1000) < 200 {
					// process the request but force discard of reply.
					c1 := conn.(*net.UnixConn)
					f, _ := c1.File()
					err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
					if err != nil {
						fmt.Printf("shutdown: %v\n", err)
					}
					go rpcs.ServeConn(conn)
				} else {
					go rpcs.ServeConn(conn)
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && kv.dead == false {
				fmt.Printf("KVRaft(%v) accept: %v\n", me, err.Error())
				kv.kill()
			}
		}
	}()

	return kv
}
