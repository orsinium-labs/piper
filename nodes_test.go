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
	cmdNode, err := piper.CommandSource(cmd, 1024)
	if err != nil {
		t.Fatalf("create command source: %v", err)
	}
	err = cmd.Start()
	if err != nil {
		t.Fatalf("start command: %v", err)
	}
	err = piper.Wait(piper.Pipe3(
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
