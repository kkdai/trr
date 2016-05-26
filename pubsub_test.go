package tszpubsub

import (
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	tsb := NewTszPubsub(10)
	tsb.PublishTimeData("d1", 0, 0)
}

func TestBasicRead(t *testing.T) {
	tsb := NewTszPubsub(10)
	t0, _ := time.ParseInLocation("Jan _2 2006 15:04:05", "Mar 24 2015 02:00:00", time.Local)
	tunix := uint32(t0.Unix())
	tsb.PublishTimeData("d1", tunix, 12)
	tt, v, err := tsb.ReadChanTopic("d1")
	if err != nil || tt != tunix || v != float64(12) {
		t.Error("Cannot get correct value", t, v)
	}
}
