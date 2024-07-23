package circlebuff

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

type Numbers interface {
	constraints.Float | constraints.Integer
}
type CircleBuffer[T Numbers] struct {
	valuesNs []calcDataItem[T]
	maxAge   time.Duration
	len      int
	idxCurr  int
	mux      sync.RWMutex
}

var (
	ErrNoValues = errors.New("no values")
)

type calcDataItem[T Numbers] struct {
	value T
	tm    time.Time
}

func New[T Numbers](n int, maxAge time.Duration) *CircleBuffer[T] {
	if n == 0 {
		n = 1
	}

	return &CircleBuffer[T]{
		maxAge:   maxAge,
		valuesNs: newSlice[T](n, n),
		idxCurr:  -1,
	}
}

func newSlice[T Numbers](l, c int) []calcDataItem[T] {
	values := make([]calcDataItem[T], l, c)
	return values
}

func (c *CircleBuffer[T]) AddValue(v T) {
	c.AddValueWithTm(v, time.Now().UTC())
}

func (c *CircleBuffer[T]) AddValueWithTm(v T, tm time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.idxCurr++

	if c.idxCurr >= len(c.valuesNs) {
		c.idxCurr = 0
	}

	if c.len < len(c.valuesNs) {
		c.len = c.idxCurr + 1
	}

	c.valuesNs[c.idxCurr] = calcDataItem[T]{
		value: v,
		tm:    tm,
	}
}

func (c *CircleBuffer[T]) Avg() (float64, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	var sum T
	var cnt int

	for i := 0; i < c.len; i++ {
		if time.Since(c.valuesNs[i].tm) > c.maxAge {
			continue
		}

		sum += c.valuesNs[i].value
		cnt++
	}

	if cnt == 0 {
		return 0, ErrNoValues
	}

	return float64(sum) / float64(cnt), nil
}

func (c *CircleBuffer[T]) Sum() (T, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	var sum T
	var cnt int

	for i := 0; i < c.len; i++ {
		if time.Since(c.valuesNs[i].tm) > c.maxAge {
			continue
		}

		sum += c.valuesNs[i].value
		cnt++
	}

	if cnt == 0 {
		return 0, ErrNoValues
	}

	return sum, nil
}
