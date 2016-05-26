package tszpubsub

import (
	tsz "github.com/dgryski/go-tsz"
)

type tszChan struct {
	timeSeries tsz.Series
	topic      string
	channel    chan interface{}
}

//TszPubsub :
type TszPubsub struct {
	topic     []string
	chanToTsz []tszChan
}

//NewTszPubsub :
func NewTszPubsub() *TszPubsub {
	return new(TszPubsub)
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
		newTsz := tszChan{topic: topic, channel: make(chan interface{})}
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
func (t *TszPubsub) ReadChanTopic(topic string) (uint32, float64) {

	return 0, 0
}

func (t *TszPubsub) loop() {

}
