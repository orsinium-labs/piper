package piper

import (
	"context"
	"errors"
	"sync"
)

type node interface {
	Run(context.Context, *sync.WaitGroup, chan<- error)
}

// Run the pipeline.
//
// If context is cancelled, all the nodes are cancelled
// ([NodeContext.Send] and [NodeContext.Recv] will return false).
//
// Any errors returned by node handlers or emitted using [NodeContext.Error]
// are emitted into the returned channel.
// The channel is closed when all nodes exit.
func Run(ctx context.Context, nodes ...node) <-chan error {
	errors := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go node.Run(ctx, &wg, errors)
	}
	go func() {
		wg.Wait()
		// If context is canceled, emit that as an error.
		// However, make sure to not block if there is nobody reading errors.
		select {
		case <-ctx.Done():
			select {
			case errors <- ctx.Err():
			default:
			}
		default:
		}
		close(errors)
	}()
	return errors
}

// Wrap [Run], wait for all nodes to finish, return combined errors if any.
func Wait(errs <-chan error) error {
	var result []error
	for err := range errs {
		result = append(result, err)
	}
	if len(result) == 0 {
		return nil
	}
	if len(result) == 1 {
		return result[0]
	}
	return errors.Join(result...)
}
