package monitor

import (
	"log"
	"sync/atomic"
	"time"
)

type Monitor struct {
	success uint64
}

func (m *Monitor) Success() {
	atomic.AddUint64(&m.success, 1)
}

func (m *Monitor) Start(d time.Duration) {
	t := time.NewTicker(d)
	start := time.Now()
	pres := uint64(0)
	for _ = range t.C {
		s := atomic.LoadUint64(&m.success)
		totalSpeed := float64(s) / time.Now().Sub(start).Seconds()
		speed := float64(s-pres) / d.Seconds()
		log.Printf("[stat] [success:%d] [speed:%.2f] [total_speed:%.2f]\n", s, speed, totalSpeed)
		pres = s
	}
}
