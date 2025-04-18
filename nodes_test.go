package piper_test

import (
	"os/exec"
	"strconv"
	"testing"

	"github.com/orsinium-labs/piper"
)

func TestCommandSource(t *testing.T) {
	sum := 0
	reduce := piper.Each(func(x []byte) error {
		i, err := strconv.ParseInt(string(x), 10, 32)
		if err != nil {
			t.Fatalf("parse int: %v", err)
		}
		sum += int(i)
		return nil
	})
	cmd := exec.Command("seq", "5", "11")
	cmdNode := piper.CommandSource(cmd, 1024)
	err := piper.Wait(piper.Pipe3(
		t.Context(),
		cmdNode,
		piper.BytesLinesNode(),
		reduce,
	))
	if err != nil {
		t.Fatal(err)
	}
	if sum != 5+6+7+8+9+10+11 {
		t.Fatalf("sum: %d", sum)
	}
}

func TestCommandSink(t *testing.T) {
	gen := piper.NewNode(func(nc *piper.NodeContext[struct{}, []byte]) error {
		ok := nc.Send([]byte("hello"))
		if !ok {
			t.Fatal("should be ok")
		}
		ok = nc.Send([]byte("world"))
		if !ok {
			t.Fatal("should be ok")
		}
		return nil
	})
	cmd := exec.Command("rev")
	cmdNode := piper.CommandSink(cmd)
	err := piper.Wait(piper.Pipe2(
		t.Context(),
		gen,
		cmdNode,
	))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCommandNode(t *testing.T) {
	gen := piper.NewNode(func(nc *piper.NodeContext[struct{}, []byte]) error {
		ok := nc.Send([]byte("hello\n"))
		if !ok {
			t.Fatal("should be ok")
		}
		ok = nc.Send([]byte("world\n"))
		if !ok {
			t.Fatal("should be ok")
		}
		return nil
	})
	cmd := exec.Command("rev")
	cmdNode := piper.CommandNode(cmd, 1024)
	sum := []byte{}
	reduce := piper.Each(func(x []byte) error {
		sum = append(sum, '|')
		sum = append(sum, x...)
		return nil
	})
	err := piper.Wait(piper.Pipe4(
		t.Context(),
		gen,
		cmdNode,
		piper.BytesLinesNode(),
		reduce,
	))
	if err != nil {
		t.Fatal(err)
	}
	act := string(sum)
	if act != "|olleh|dlrow" {
		t.Fatal(act)
	}
}
