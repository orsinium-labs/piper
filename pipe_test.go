package pipe_test

import (
	"testing"

	"github.com/ourmindbv/ourmind-peach/peach/pipe"
)

func TestPipe1Noop(t *testing.T) {
	n1 := pipe.NewNode(func(nc *pipe.NodeContext[struct{}, int]) error {
		return nil
	})
	pipe.Run(t.Context(), n1)
}
