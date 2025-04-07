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
	ctx    context.Context
	in     *wireIn[I]
	out    *wireOut[O]
	errors chan<- error
}

func (n NodeContext[I, O]) Context() context.Context {
	return n.ctx
}

// Read a message from the node input.
func (n NodeContext[I, O]) Recv() (I, bool) {
	select {
	case data, more := <-n.in.ch:
		if !more {
			var def I
			return def, false
		}
		return data, true
	case <-n.ctx.Done():
		var def I
		return def, false
	}
}

// Write a message to the node output.
func (n NodeContext[I, O]) Send(data O) bool {
	select {
	case n.out.ch <- data:
		return true
	case <-n.out.done:
		// The consumer is dead, no need to send anything anymore.
		return false
	case <-n.ctx.Done():
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
	case <-n.ctx.Done():
		return false
	}
}
