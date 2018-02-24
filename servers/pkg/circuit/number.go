package circuit

import (
	"sync"
	"sync/atomic"
	"time"
)

// SecondBucket tracks a numberBucket over a bounded number of
// time buckets. Currently the buckets are one second long and only the last 10 seconds are kept.
type SecondBucket struct {
	Buckets sync.Map
	timeout int64
}

type value struct {
	Value uint64
}

// NewSecondBucket initializes a RollingNumber struct.
func NewSecondBucket() *SecondBucket {
	return &SecondBucket{timeout: 10}
}

func (r *SecondBucket) getCurrentBucket() *value {
	now := time.Now().Unix()
	v, b := r.Buckets.LoadOrStore(now, &value{})
	if b {
		r.removeOldBuckets()
	}
	return v.(*value)
}

func (r *SecondBucket) removeOldBuckets() {
	before := time.Now().Unix() - r.timeout
	r.Buckets.Range(func(k, v interface{}) bool {
		if i, ok := k.(int64); ok && i < before {
			r.Buckets.Delete(k)
		}
		return true
	})
}

// Increment increments the number in current SecondBucket.
func (r *SecondBucket) Increment(i uint64) {
	if i == 0 {
		return
	}
	b := r.getCurrentBucket()
	atomic.AddUint64(&b.Value, i)
}

// UpdateMax updates the maximum value in the current bucket.
func (r *SecondBucket) UpdateMax(n uint64) {
	b := r.getCurrentBucket()
	if n > b.Value {
		atomic.SwapUint64(&b.Value, n)
	}
}

// Sum sums the values over the buckets in the last 10 seconds.
func (r *SecondBucket) Sum(now time.Time) uint64 {
	sum := uint64(0)
	last := now.Unix() - r.timeout
	r.Buckets.Range(func(k, v interface{}) bool {
		if t, ok := k.(int64); ok && t >= last {
			sum += v.(*value).Value
		}
		return true
	})
	return sum
}

// Max returns the maximum value seen in the last 10 seconds.
func (r *SecondBucket) Max(now time.Time) uint64 {
	var max uint64
	last := now.Unix() - r.timeout
	r.Buckets.Range(func(k, v interface{}) bool {
		if t, ok := k.(int64); ok && t > last && v.(*value).Value > max {
			max = v.(*value).Value
		}
		return true
	})
	return max
}

//Average 指定一段时间内的平均值
func (r *SecondBucket) Average() float64 {
	return float64(r.Sum(time.Now())) / float64(r.timeout)
}
