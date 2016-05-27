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
// limitations under the License

package trr

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestClientWithSingleServerWithRawData(t *testing.T) {
	srv := StartServer("127.0.0.1:1234", 1)

	client := MakeClerk("127.0.0.1:1234")
	client.putRaw("t1", []byte("v1"))
	ret := client.getRaw("t1")

	log.Println("got:", ret)
	if string(ret) != "v1" {
		t.Error("Client get error:", ret)
	}

	srv.kill()
	os.RemoveAll("trr-1")
}

func TestClientWithSingleServerWithTimeDataFirst(t *testing.T) {
	t0, _ := time.ParseInLocation("Jan _2 2006 15:04:05", "Mar 24 2015 02:00:00", time.Local)
	tunix := uint32(t0.Unix())

	srv := StartServer("127.0.0.1:1239", 1)

	client := MakeClerk("127.0.0.1:1239")
	client.PutTimeData("t1", tunix, 10)
	tt, vv, err := client.GetTimeData("t1")
	if err != nil || tt != tunix || vv != 10 {
		t.Error("Simple time get error", tt, vv, err)
	}

	tt, vv, err = client.GetTimeData("t1")
	if err == nil {
		t.Error("Should be error when no value", err)
	}

	srv.kill()
	os.RemoveAll("trr-1")
}

func TestClientWithSingleServerWithTimeDataSecond(t *testing.T) {
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
		t.Error("Simple time get error", tt, vv, err)
	}

	tt, vv, err = client.GetTimeData("t1")
	if err != nil || tt != t1unix || vv != 12 {
		t.Error("Simple time get error", tt, vv, err)
	}

	tt, vv, err = client.GetTimeData("t2")
	if err == nil {
		t.Error("Should be error when no value", err)
	}

	tt, vv, err = client.GetTimeData("t1")
	if err != nil || tt != t2unix || vv != 14 {
		t.Error("Simple time get error", tt, vv, err)
	}

	srv.kill()
	os.RemoveAll("trr-1")
}
