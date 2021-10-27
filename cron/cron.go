package cron

import (
	"log"
	"time"
)

const INTERVAL_PERIOD time.Duration = 24 * time.Hour

const HOUR_TO_TICK int = 11
const MINUTE_TO_TICK int = 55
const SECOND_TO_TICK int = 00

type jobTicker struct {
	timer *time.Timer
	f     func() error
}

func Run(f func() error) {
	jobTicker := &jobTicker{
		f: f,
	}
	jobTicker.updateTimer()
	for {
		<-jobTicker.timer.C
		log.Println("Running cron task")
		jobTicker.f()
		jobTicker.updateTimer()
	}
}

func (t *jobTicker) updateTimer() {
	nextTick := time.Date(time.Now().Year(), time.Now().Month(),
		time.Now().Day(), HOUR_TO_TICK, MINUTE_TO_TICK, SECOND_TO_TICK, 0, time.Local)
	if !nextTick.After(time.Now()) {
		nextTick = nextTick.Add(INTERVAL_PERIOD)
	}
	log.Println(" cront schedulted at :", nextTick)
	diff := nextTick.Sub(time.Now())
	if t.timer == nil {
		t.timer = time.NewTimer(diff)
	} else {
		t.timer.Reset(diff)
	}
}
