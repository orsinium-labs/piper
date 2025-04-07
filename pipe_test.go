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
	if err.Error() != "oh no!" {
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
	doubler := piper.NewNode(func(nc *piper.NodeContext[int, int]) error {
		for n := range nc.Iter() {
			ok := nc.Send(n)
			if !ok {
				break
			}
		}
		return nil
	})
	sum := 0
	summer := piper.NewNode(func(nc *piper.NodeContext[int, struct{}]) error {
		for n := range nc.Iter() {
			sum += n
		}
		return nil
	})

	piper.Connect(numbers, doubler)
	piper.Connect(doubler, summer)
	err := piper.Wait(piper.Run(t.Context(), numbers, doubler, summer))
	if err != nil {
		t.Fatal(err)
	}
	if sum != 3+4+5 {
		t.Fatal(sum)
	}
}
