package piper

import (
	"context"
	"sync"
)

type Node[I, O any] struct {
	context *NodeContext[I, O]
	handler func(*NodeContext[I, O]) error
}

func NewNode[I, O any](h func(*NodeContext[I, O]) error) *Node[I, O] {
	return &Node[I, O]{
		context: &NodeContext[I, O]{
			In:  &wireIn[I]{},
			Out: &wireOut[O]{},
		},
		handler: h,
	}
}

// Connect two nodes together.
func Connect[T, X, Y any](n1 *Node[X, T], n2 *Node[T, Y]) {
	ch := make(chan T)
	n1.context.Out.ch = ch
	n2.context.In.ch = ch

	done := make(chan struct{})
	n1.context.Out.done = done
	n2.context.In.done = done
}

func (n *Node[I, O]) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	errors chan<- error,
) {
	n.context.Ctx = ctx
	defer func() {
		wg.Done()
		if n.context.Out.ch != nil {
			close(n.context.Out.ch)
		}
		if n.context.In.done != nil {
			close(n.context.In.done)
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
