package main
import "flag"
import "fmt"
import "time"
import "leb.io/limiter"
import "github.com/tildeleb/cuckoo/siginfo"

var procs = flag.Int("p", 10, "goroutines")
var iter = flag.Int("i", 10, "iterations")
var nlocks = flag.Int("l", 1, "locks")
var freq = flag.Duration("f", 1.0 * time.Second, "freqencies")
var maxCalls = flag.Int("m", 10, "max calls")
var tp = flag.Duration("t", 1.0 * time.Second, "time period")
var vf = flag.Bool("v", false, "verbose")

var counters []int
var term chan bool		// channel used for termination

func tdiff(begin, end time.Time) time.Duration {
    d := end.Sub(begin)
    return d
}

func apiCaller(t *limiter.Throttle, rate time.Duration, inst, n int) {
	//fmt.Printf("%d: apiCaller: inst, rate=%v, n=%d\n", inst, rate, n)
	ticker := time.NewTicker(rate)
	for i := 0; i < n; i++ {
		<- ticker.C
		t.Limit(inst)
		counters[inst]++
		if *vf {
			fmt.Printf("%d ", inst)
		}
	}
	term <- true
}

func f() {
	for k, v := range counters {
		fmt.Printf("counters[%d]=%d\n", k, v)
	}
}

func main() {
	flag.Parse()
	fmt.Printf("Throttle: tp=%v, freq=%v\n", *tp, *freq)
	counters = make([]int, *procs)
	term = make(chan bool)
    siginfo.SetHandler(f)
	t := limiter.New(*maxCalls, *tp, *procs)
	for i:= 0; i < *procs; i++ {
		go apiCaller(t, *freq, i, *iter)
	}
	// and the apiCallers run
	tcnt := *procs
	first := time.Now()
	for {
		select {
		case <-term:
			if tcnt == *procs {
				first = time.Now()
			}
			tcnt--
			//fmt.Printf("tcnt=%d\n", tcnt)
		}
		if tcnt == 0 {
			break
		}
	}
	stop := time.Now()
	skew := tdiff(first, stop)
	fmt.Printf("skew=%v\n", skew)
}