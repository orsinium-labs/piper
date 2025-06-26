package piper

import (
	"context"
	"fmt"
	"iter"
	"sync/atomic"
)

type NodeState uint8

const (
	// The node hasn't called Recv or Send yet.
	NodeStateNew NodeState = 0
	// The node has called Recv and is waiting for incoming messages.
	NodeStateRecv NodeState = 1
	// The node has received a message and now handles it.
	NodeStateProcess NodeState = 2
	// The node has called Send and is waiting for the next step to consume the message.
	NodeStateSend NodeState = 3
	// The node has sent a message with Send but hasn't called Recv yet.
	NodeStateIdle NodeState = 4
	// The node handler has exited without an error.
	NodeStateDone NodeState = 5
	// The node handler has exited with an error.
	NodeStateFailed NodeState = 6
)

type wireIn[T any] struct {
	ch   <-chan T
	done chan<- struct{}
}

type wireOut[T any] struct {
	// Closed by writer when the writer exits.
	ch chan<- T
	// Closed by reader when the reader exits.
	done <-chan struct{}
}

type NodeContext[I, O any] struct {
	ctx    context.Context
	in     *wireIn[I]
	out    *wireOut[O]
	errors chan<- error
	name   string
	index  int
	state  *int32
}

// Get the context passed into [Run].
func (n NodeContext[I, O]) Context() context.Context {
	return n.ctx
}

// Read a message from the node input.
//
// Returns false if the pipeline is cancelled
// or if the input node has exited and will produce no more messages.
func (n NodeContext[I, O]) Recv() (I, bool) {
	n.setState(NodeStateRecv)
	select {
	case data, more := <-n.in.ch:
		n.setState(NodeStateProcess)
		if !more {
			var def I
			return def, false
		}
		return data, true
	case <-n.ctx.Done():
		n.setState(NodeStateProcess)
		var def I
		return def, false
	}
}

// Write a message to the node output.
//
// Returns false if the pipeline is cancelled
// or the consumer node has exited and cannot handle messages.
func (n NodeContext[I, O]) Send(data O) bool {
	n.setState(NodeStateSend)
	select {
	case n.out.ch <- data:
		n.setState(NodeStateIdle)
		return true
	case <-n.out.done:
		n.setState(NodeStateIdle)
		// The consumer is dead, no need to send anything anymore.
		return false
	case <-n.ctx.Done():
		n.setState(NodeStateIdle)
		return false
	}
}

// Iterate over input messages.
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

// Same as calling [fmt.Errorf] and then [NodeContext.Error].
//
// Returns false if the pipeline is cancelled.
func (n NodeContext[I, O]) Errorf(format string, a ...any) bool {
	return n.Error(fmt.Errorf(format, a...))
}

// Emit an error without interrupting the pipeline.
//
// Returns false if the pipeline is cancelled.
func (n NodeContext[I, O]) Error(err error) bool {
	if err == nil {
		return !n.Cancelled()
	}
	if n.name != "" {
		err = fmt.Errorf("node %s: %w", n.name, err)
	} else {
		err = fmt.Errorf("node #%d: %w", n.index, err)
	}
	select {
	case n.errors <- err:
		return true
	case <-n.ctx.Done():
		return false
	}
}

// Returns true if the pipeline's input context is done.
func (n NodeContext[I, O]) Cancelled() bool {
	select {
	case <-n.ctx.Done():
		return true
	default:
		return false
	}
}

func (n NodeContext[I, O]) setState(s NodeState) {
	atomic.StoreInt32(n.state, int32(s))
}

type ctxKey[T any] struct{}

// Attach the given value to the context.
//
// Must be called on the context before passing it into Pipe or [Run].
//
// The context can store only one value of the given type.
// Panics if there is already a value of the same type in the context.
func With[T any](ctx context.Context, val T) context.Context {
	key := ctxKey[T]{}
	if ctx.Value(key) != nil {
		panic("context already contains value of the given type")
	}
	ctx = context.WithValue(ctx, key, val)
	return ctx
}

// Get from the node context the value added using [With].
//
// Panics if there is no value of the given type in the context.
func Get[T, I, O any](nc *NodeContext[I, O]) T {
	raw := nc.ctx.Value(ctxKey[T]{})
	if raw == nil {
		panic("no value of the given type in the context")
	}
	return raw.(T)
}
