package piper_test

import (
	"context"
	"errors"
	"testing"

	"github.com/orsinium-labs/piper"
)

func TestPipe1Noop(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		return nil
	})
	err := piper.Wait(piper.Run(t.Context(), n1))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPipe1Cancel(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		ok := nc.Send(struct{}{})
		if ok {
			t.Fatalf("send should have failed")
		}
		return nil
	})
	ctx, cancel := context.WithCancel(t.Context())
	errs := piper.Run(ctx, n1)
	cancel()
	err := piper.Wait(errs)
	if err != ctx.Err() {
		t.Fatal(err)
	}
}

func TestPipe1Error(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		return errors.New("oh no!")
	})
	err := piper.Wait(piper.Run(t.Context(), n1))
	if err.Error() != "node #1: exited with error: oh no!" {
		t.Fatal(err)
	}
}

func TestPipe1Errorf(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		nc.Errorf("well: %v", "damn")
		return nil
	})
	err := piper.Wait(piper.Run(t.Context(), n1))
	if err.Error() != "node #1: well: damn" {
		t.Fatal(err)
	}
}

func TestPipe1Cencel(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		ok := nc.Send(struct{}{})
		if ok {
			t.Fatal("expected Send to fail")
		}
		_, ok = nc.Recv()
		if ok {
			t.Fatal("expected Recv to fail")
		}
		if !nc.Cancelled() {
			t.Fatal("expected Cancelled to be true")
		}
		return nil
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := piper.Wait(piper.Run(ctx, n1))
	if err.Error() != "context canceled" {
		t.Fatal(err)
	}
}

func TestPipe1Name(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		return errors.New("oh no!")
	}).WithName("hello")
	err := piper.Wait(piper.Run(t.Context(), n1))
	if err.Error() != "node hello: exited with error: oh no!" {
		t.Fatal(err)
	}
}

func TestWith(t *testing.T) {
	n1 := piper.NewNode(func(nc *piper.NodeContext[struct{}, struct{}]) error {
		name := piper.Get[string](nc)
		if name != "aragorn" {
			t.Fatalf("expected aragorn, got %s", name)
		}
		return nil
	})
	ctx := piper.With(t.Context(), "aragorn")
	err := piper.Wait(piper.Run(ctx, n1))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPipe3(t *testing.T) {
	numbers := piper.NewNode(func(nc *piper.NodeContext[struct{}, int]) error {
		for i := 3; i < 6; i++ {
			ok := nc.Send(i)
			if !ok {
				t.Fatal("not ok")
			}
		}
		return nil
	})
	doubler := piper.Map(func(n int) (int, error) {
		return n * 2, nil
	})
	sum := 0
	summer := piper.Each(func(n int) error {
		sum += n
		return nil
	})

	piper.Connect(numbers, doubler)
	piper.Connect(doubler, summer)
	err := piper.Wait(piper.Run(t.Context(), numbers, doubler, summer))
	if err != nil {
		t.Fatal(err)
	}
	if sum != (3+4+5)*2 {
		t.Fatal(sum)
	}
}
