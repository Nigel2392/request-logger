package accumulator

import (
	"sync"
	"time"

	"github.com/Nigel2392/go-datastructures/stack"
)

// An accumulator adds a number of items to a queue and
// flushes the queue when the queue is full or a certain time has passed.
//
// The items can then be processed by a worker with the specified handler.
type Accumulator[T any] struct {
	// The queue for the items.
	Queue stack.Stack[T]

	// The number of items to accumulate before the queue is flushed.
	FlushSize int

	// The time to wait before flushing the queue.
	FlushInterval time.Duration

	// Reset the flush interval after a push.
	ResetAfterPush bool

	// ticker is a ticker which is used to flush the queue.
	ticker *time.Ticker

	// The mutex used to lock the queue.
	mutex *sync.Mutex

	// closeChan is a channel which is closed when the batch is closed.
	closeChan chan struct{}

	// The function which is called when the queue is flushed.
	FlushFunc func([]T)
}

// NewAccumulator creates a new accumulator which accumulates items and flushes them when the flush size is reached or the flush interval is reached.
func NewAccumulator[T any](flushSize int, flushInterval time.Duration, flushFunc func([]T)) *Accumulator[T] {
	var a = &Accumulator[T]{
		FlushSize:     flushSize,
		FlushInterval: flushInterval,
		FlushFunc:     flushFunc,
		Queue:         stack.Stack[T]{},
		mutex:         &sync.Mutex{},
		closeChan:     make(chan struct{}),
	}
	a.ticker = time.NewTicker(flushInterval)
	go a.worker()
	return a
}

func (a *Accumulator[T]) worker() {
	for {
		select {
		case <-a.closeChan:
			return
		case <-a.ticker.C:
			a.Flush()
		default:
			a.mutex.Lock()
			var needsFlush bool = a.Queue.Len() >= a.FlushSize
			if needsFlush {
				a.Flush()
			}
			a.mutex.Unlock()
			time.Sleep(a.FlushInterval / 10)
		}
	}
}

// Push adds an item to the accumulator.
func (a *Accumulator[T]) Push(item T) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.Queue.Push(item)
	var needsFlush bool = a.Queue.Len() >= a.FlushSize
	if needsFlush {
		a.Flush()
	}
	if a.ResetAfterPush {
		a.ticker.Reset(a.FlushInterval)
	}
}

// Flush flushes the queue.
func (a *Accumulator[T]) Flush() {
	var lockedHere = a.mutex.TryLock()
	if lockedHere {
		defer a.mutex.Unlock()
	}
	var items = make([]T, 0, a.Queue.Len())
	for {
		item, ok := a.Queue.PopOK()
		if !ok {
			break
		}
		items = append(items, item)
	}
	if len(items) > 0 {
		a.FlushFunc(items)
	}
}

// Close closes the accumulator.
func (a *Accumulator[T]) Close() {
	a.Flush()
	close(a.closeChan)
}
