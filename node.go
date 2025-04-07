package piper

import (
	"context"
	"sync"
)

// A unit of work, a background process reading input, doing work, and writing to output.
type Node[I, O any] struct {
	context *NodeContext[I, O]
	handler func(*NodeContext[I, O]) error
}

func NewNode[I, O any](h func(*NodeContext[I, O]) error) *Node[I, O] {
	return &Node[I, O]{
		context: &NodeContext[I, O]{
			in:  &wireIn[I]{},
			out: &wireOut[O]{},
		},
		handler: h,
	}
}

// Run the node. Don't call directly, use [Run] instead.
func (n *Node[I, O]) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	errors chan<- error,
) {
	n.context.ctx = ctx
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
		select {
		case errors <- err:
		case <-ctx.Done():
		}
	}
}
