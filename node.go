package piper

import (
	"context"
	"fmt"
	"sync"
)

// A unit of work, a background process reading input, doing work, and writing to output.
type Node[I, O any] struct {
	context *NodeContext[I, O]
	handler func(*NodeContext[I, O]) error
}

func NewNode[I, O any](h func(*NodeContext[I, O]) error) *Node[I, O] {
	if h == nil {
		panic("node handler must be non-nil")
	}
	return &Node[I, O]{
		context: &NodeContext[I, O]{
			in:  &wireIn[I]{},
			out: &wireOut[O]{},
		},
		handler: h,
	}
}

// Set the node name.
//
// If set, it will be added to all errors emitted by the node.
func (n *Node[I, O]) WithName(name string) *Node[I, O] {
	n.context.name = name
	return n
}

// Catch panics and transform them into errors using the given handler.
//
// If the given panic handler is nil, [fmt.Errorf] will be used.
func (n *Node[I, O]) WithPanicHandler(ph func(any) error) *Node[I, O] {
	origHandler := n.handler
	if ph == nil {
		ph = func(p any) error {
			return fmt.Errorf("panic: %v", p)
		}
	}
	n.handler = func(nc *NodeContext[I, O]) (err error) {
		defer func() {
			p := recover()
			if p != nil {
				err = ph(p)
			}
		}()
		return origHandler(nc)
	}
	return n
}

// Run the node. Don't call directly, use [Run] instead.
func (n *Node[I, O]) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	errors chan<- error,
	index int,
) {
	n.context.ctx = ctx
	n.context.index = index
	n.context.errors = errors
	defer func() {
		wg.Done()
		if n.context.out.ch != nil {
			close(n.context.out.ch)
		}
		if n.context.in.done != nil {
			close(n.context.in.done)
		}
	}()
	err := n.handler(n.context)
	if err != nil {
		n.context.Errorf("exited with error: %w", err)
	}
}
