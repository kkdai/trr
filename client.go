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
	"net/rpc"

	tsz "github.com/dgryski/go-tsz"
)

type timeData struct {
	localTimeSeries *tsz.Series
	localIter       *tsz.Iter
}

//Clerk :
type Clerk struct {
	server string
	db     map[string]timeData
}

//MakeClerk :
func MakeClerk(server string) *Clerk {
	ck := new(Clerk)
	ck.server = server
	ck.db = make(map[string]timeData)
	return ck
}

func call(srv string, rpcname string,
	args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("tcp", srv)
	if errx != nil {
		log.Println("[Client] Dial err:", errx)
		return false
	}
	defer c.Close()

	err := c.Call(rpcname, args, reply)
	if err != nil {
		log.Println("[Client] Call err:", err)
		return true
	}

	fmt.Println(err)
	return false
}

//Get
// fetch the current value for a key.
func (ck *Clerk) getRaw(key string) []byte {
	arg := GetArgs{Key: key}
	var reply GetReply
	err := call(ck.server, "KVRaft.Get", &arg, &reply)
	if err {
		log.Println(reply.Err)
	}

	return reply.Value
}

//GetTimeData :
func (ck *Clerk) GetTimeData(key string) (uint32, float64, error) {

	vT, exist := ck.db[key]
	if !exist || vT.localIter == nil || vT.localIter.Next() == false {
		timeData := ck.getRaw(key)
		if timeData == nil {
			return 0, 0, errors.New("No key")
		}
		var err error
		vT.localIter, err = tsz.NewIterator(timeData)
		if err != nil {
			return 0, 0, errors.New("No value")
		}
		vT.localIter.Next()
		ck.db[key] = vT
	}

	tt, vv := ck.db[key].localIter.Values()
	return tt, vv, nil
}

//putRaw :
func (ck *Clerk) putRaw(key string, value []byte) {
	arg := PutArgs{Key: key, Value: value}
	var reply PutReply

	err := call(ck.server, "KVRaft.Put", &arg, &reply)
	if err {
		log.Println(reply.Err)
	}
}

//PutTimeData :
func (ck *Clerk) PutTimeData(key string, time uint32, value float64) {
	if _, exist := ck.db[key]; !exist {
		newT := timeData{}
		newT.localTimeSeries = tsz.New(time)
		ck.db[key] = newT
	}

	ck.db[key].localTimeSeries.Push(time, value)
	newT := timeData{localTimeSeries: ck.db[key].localTimeSeries, localIter: ck.db[key].localTimeSeries.Iter()}
	ck.db[key] = newT
}

//PutTimeDataBack :
func (ck *Clerk) PutTimeDataBack(key string, time uint32, value float64) {
	ck.PutTimeData(key, time, value)

	ck.db[key].localTimeSeries.Finish()
	allValues := ck.db[key].localTimeSeries.Bytes()
	ck.putRaw(key, allValues)
	delete(ck.db, key)
}
