package pipe

import (
	"context"
	"sync"
)

type node interface {
	Run(context.Context, *sync.WaitGroup, chan<- error)
}

func Run(ctx context.Context, nodes ...node) <-chan error {
	errors := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go node.Run(ctx, &wg, errors)
	}
	go func() {
		wg.Wait()
		close(errors)
	}()
	return errors
}

func Wait(<-chan error) error {
	// ...
}

type WireIn[T any] struct {
	ch   <-chan T
	done chan<- struct{}
}

type WireOut[T any] struct {
	// Closed by writer when the writer exits.
	ch chan<- T
	// Closed by reader when the reader exits
	done <-chan struct{}
}

type NodeContext[I, O any] struct {
	Ctx context.Context
	In  *WireIn[I]
	Out *WireOut[O]
}

type Node[I, O any] struct {
	context *NodeContext[I, O]
	handler func(*NodeContext[I, O]) error
}

func NewNode[I, O any](h func(*NodeContext[I, O]) error) *Node[I, O] {
	return &Node[I, O]{
		context: &NodeContext[I, O]{},
		handler: h,
	}
}

func Connect[T, X, Y any](n1 *Node[X, T], n2 *Node[T, Y]) {
	ch := make(chan T)
	n1.context.Out.ch = ch
	n2.context.In.ch = ch

	done := make(chan struct{})
	n1.context.Out.done = done
	n2.context.In.done = done
}

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

func (n *Node[I, O]) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	errors chan<- error,
) {
	n.context.Ctx = ctx
	defer func() {
		wg.Done()
		close(n.context.Out.ch)
		close(n.context.In.done)
	}()
	err := n.handler(n.context)
	if err != nil {
		select {
		case errors <- err:
		case <-ctx.Done():
		}
	}
}
