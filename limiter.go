// Copyright © 2015 Lawrence E. Bakst. All rights reserved.

// Package limiter implements an API rate limiter which can be used on the client or server side.
// Limiter works to prevent livelock by waking waiting goroutines in FIFO order.
// The package is safe and can be used from goroutines, no wrapper mutex is required.
package limiter

import (
	"time"
	"fmt"
	"sync/atomic"
)

type Throttle struct {
	instances	int				// number of unique goroutines using the limiter
	calls		int32			// limit to max of calls per dur	
	dur			time.Duration	// max calls per dur		
	counter		int32			// atomic counter of calls, counts up, can be greater than max
	ticker		*time.Ticker	// ticks by dur
	tim			time.Time		// for debugging
	sleepers	chan int		// channel for notifcation of sleeping goroutines, they write their instance
	wakeups		[]chan struct{}	// waiting goroutines, one channel for each instance
}

// This function limits the number of API calls.
func (t *Throttle) limiter() {
	var cnt int
	for {
//top:
		//fmt.Printf("limiter: Set, counter=%d\n", 0 + int32(cnt))
		atomic.StoreInt32(&t.counter, 0 + int32(cnt))
		for i := 0; i < cnt; i++ {
			t.wakeups[<-t.sleepers] <-struct{}{}
		}
		t.tim = <- t.ticker.C
//		atomic.StoreInt32(&t.Counter, 0)
		//fmt.Printf("limiter: tim=%v, counter=%d\n", t.Tim, t.Counter)
/*
		cnt = 0
		for {
			select {
			case <-t.C:
				cnt++ // // count each waiting process
			default:
				fmt.Printf("limiter: breakout cnt=%d\n", cnt)
				goto top
			}
		}
*/
		cnt = len(t.sleepers)
		//fmt.Printf("limiter: breakout cnt=%d\n", cnt)
	}
}

// Create a throttle structure that limits API calling to calls per dur.
// Each goroutine calling Limit must call it with a unique number between 0 and instamces-1.
func New(calls int, dur time.Duration, instances int) *Throttle {
	fmt.Printf("maxCalls: %d, sec=%v\n", calls, dur)
	t := new(Throttle)
	t.calls = int32(calls)
	t.dur = dur
	t.instances = instances
	t.ticker = time.NewTicker(dur)
	t.sleepers = make(chan int, instances)
	t.wakeups = make([]chan struct{}, instances)
	for i := 0; i < instances; i++ {
		t.wakeups[i] = make(chan struct{})
	}
	go t.limiter()
	return t
}

// Call Limit before each API call to rate limit.
// Check increments the counter and if it is above to max it sleeps the goroutine
// until the limiter wakes it up.
func (t *Throttle) Limit(inst int) {
	if inst > t.instances-1 {
		panic("Limit")
	}
	v := atomic.AddInt32(&t.counter, 1)
	if v > t.calls {
		t.sleepers <- inst // struct{}{}
		//fmt.Printf("%d: Check: v=%d, Sleep\n", inst, v)
		<- t.wakeups[inst]
		//fmt.Printf("%d: Check: v=%d, Wake\n", inst, t.Counter)
	}
}