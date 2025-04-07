package piper

import (
	"context"
	"iter"
)

type wireIn[T any] struct {
	ch   <-chan T
	done chan<- struct{}
}

type wireOut[T any] struct {
	// Closed by writer when the writer exits.
	ch chan<- T
	// Closed by reader when the reader exits
	done <-chan struct{}
}

type NodeContext[I, O any] struct {
	Ctx    context.Context
	In     *wireIn[I]
	Out    *wireOut[O]
	errors chan<- error
}

// Read a message from the node input.
func (n NodeContext[I, O]) Recv() (I, bool) {
	select {
	case data, more := <-n.In.ch:
		if !more {
			var def I
			return def, false
		}
		return data, true
	case <-n.Ctx.Done():
		var def I
		return def, false
	}
}

// Write a message to the node output.
func (n NodeContext[I, O]) Send(data O) bool {
	select {
	case n.Out.ch <- data:
		return true
	case <-n.Out.done:
		// The consumer is dead, no need to send anything anymore.
		return false
	case <-n.Ctx.Done():
		return false
	}
}

func (n NodeContext[I, O]) Iter() iter.Seq[I] {
	return func(yield func(I) bool) {
		for {
			data, more := n.Recv()
			if !more {
				return
			}
			more = yield(data)
			if !more {
				return
			}
		}
	}
}

// Emit an error.
func (n NodeContext[I, O]) Error(err error) bool {
	select {
	case n.errors <- err:
		return true
	case <-n.Ctx.Done():
		return false
	}
}
