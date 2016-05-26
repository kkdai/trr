package tszpubsub

import (
	"testing"
)

func TestBasic(t *testing.T) {
	tsb := NewTszPubsub()
	tsb.PublishTimeData("d1", 0, 0)
	tsb.PublishTimeData("d1", 0, 0)
}

func TestBasicRead(t *testing.T) {
	tsb := NewTszPubsub()
	tsb.PublishTimeData("d1", 3122, 11102)
	tt, v, err := tsb.ReadChanTopic("d1")
	if err != nil || tt != uint32(3122) || v != float64(11102) {
		t.Error("Cannot get correct value", t, v)
	}

}
