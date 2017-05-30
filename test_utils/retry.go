package test_utils

import (
	"math"
	"time"
)

// retry for datastore's eventual consistency
func RetryWith(max int, impl func() func()) {
	for i := 0; i < max+1; i++ {
		f := impl()
		if f == nil {
			return
		}
		if i == max {
			f()
		} else {
			// Exponential backoff
			d := time.Duration(math.Pow(2.0, float64(i)) * 5.0)
			time.Sleep(d * time.Millisecond)
		}
	}
}
