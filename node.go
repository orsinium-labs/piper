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
			in:  &wireIn[I]{},
			out: &wireOut[O]{},
		},
		handler: h,
	}
}

// Connect two nodes together.
func Connect[T, X, Y any](
	n1 *Node[X, T],
	n2 *Node[T, Y],
) {
	ch := make(chan T)
	ConnectChan(n1, n2, ch)
}

func ConnectChan[T, X, Y any](
	n1 *Node[X, T],
	n2 *Node[T, Y],
	ch chan T,
) {
	n1.context.out.ch = ch
	n2.context.in.ch = ch

	done := make(chan struct{})
	n1.context.out.done = done
	n2.context.in.done = done
}

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
