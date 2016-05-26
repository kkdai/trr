package tszpubsub

import (
	"errors"
	"log"

	tsz "github.com/dgryski/go-tsz"
)

type tszChan struct {
	timeSeries *tsz.Series
	topic      string
	channel    chan interface{}
}

//TszPubsub :
type TszPubsub struct {
	topic     []string
	chanToTsz []tszChan
	capacity  int
}

//NewTszPubsub :
func NewTszPubsub(cap int) *TszPubsub {
	tt := new(TszPubsub)
	tt.capacity = cap
	go tt.loop()
	return tt
}

func isSliceContain(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

//PublishTimeData :
func (t *TszPubsub) PublishTimeData(topic string, timeData uint32, value float64) {
	if !isSliceContain(t.topic, topic) {
		t.topic = append(t.topic, topic)
		newTsz := tszChan{topic: topic, channel: make(chan interface{}, t.capacity)}
		newTsz.timeSeries = tsz.New(timeData)
		t.chanToTsz = append(t.chanToTsz, newTsz)
	}

	for k, v := range t.chanToTsz {
		if v.topic == topic {
			t.chanToTsz[k].timeSeries.Push(timeData, value)
			t.chanToTsz[k].channel <- 0
		}
	}
}

//ReadChanTopic :
func (t *TszPubsub) ReadChanTopic(topic string) (uint32, float64, error) {
	if isSliceContain(t.topic, topic) {
		for k, v := range t.chanToTsz {
			if v.topic == topic {
				<-t.chanToTsz[k].channel
				iter := t.chanToTsz[k].timeSeries.Iter()
				stillIter := iter.Next()
				if !stillIter {
					log.Println("to end")
				}
				tt, vv := iter.Values()
				return tt, vv, nil
			}
		}
	}

	return 0, 0, errors.New("Not found!")
}

func (t *TszPubsub) loop() {

}
