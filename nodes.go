package piper

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"slices"
)

// Node reading byte chunks from the command's stdout.
func CommandSource(cmd *exec.Cmd, chunkSize int) *Node[struct{}, []byte] {
	return NewNode(func(nc *NodeContext[struct{}, []byte]) (err error) {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("connect to stdout: %w", err)
		}
		defer func() {
			_ = stdout.Close()
		}()
		if cmd.Process == nil {
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("start %s: %w", cmd.Path, err)
			}
		}
		err = pipeReader(nc, stdout, chunkSize)
		if err != nil {
			return fmt.Errorf("read from stdout: %w", err)
		}
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("wait for %s: %w", cmd.Path, err)
		}
		return nil
	})
}

// Node writing byte chunks into the command's stdin.
func CommandSink(cmd *exec.Cmd) *Node[[]byte, struct{}] {
	return NewNode(func(nc *NodeContext[[]byte, struct{}]) (err error) {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("connect to stdin: %w", err)
		}
		defer func() {
			_ = stdin.Close()
		}()
		if cmd.Process == nil {
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("start %s: %w", cmd.Path, err)
			}
		}
		for chunk := range nc.Iter() {
			_, err := stdin.Write(chunk)
			if err != nil {
				return err
			}
		}
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("wait for %s: %w", cmd.Path, err)
		}
		return nil
	})
}

// A combination of [CommandSink] and [CommandSource].
//
// Run the command, write input into stdin, read output from stdout.
func CommandNode(cmd *exec.Cmd, stdoutChunkSize int) *Node[[]byte, []byte] {
	return NewNode(func(nc *NodeContext[[]byte, []byte]) (err error) {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("connect to stdout: %w", err)
		}
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("connect to stdin: %w", err)
		}
		defer func() {
			_ = stdout.Close()
			_ = stdin.Close()
		}()

		go func() {
			for chunk := range nc.Iter() {
				_, err := stdin.Write(chunk)
				if err != nil {
					nc.Errorf("write to stdin: %w", err)
				}
			}
		}()

		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("start %s: %v", cmd.Path, err)
		}
		err = pipeReader(nc, stdout, stdoutChunkSize)
		if err != nil {
			nc.Errorf("read from stdout: %w", err)
		}
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("wait for %s: %w", cmd.Path, err)
		}
		return nil
	})
}

// Node reading byte chunks from the given read-closer and closing it.
func ReadCloserSource(r io.ReadCloser, chunkSize int) *Node[struct{}, []byte] {
	return NewNode(func(nc *NodeContext[struct{}, []byte]) (err error) {
		defer func() {
			closeErr := r.Close()
			if err == nil && closeErr != nil {
				err = closeErr
			}
		}()
		return pipeReader(nc, r, chunkSize)
	})
}

// Node reading byte chunks from the given reader.
func ReaderSource(r io.Reader, chunkSize int) *Node[struct{}, []byte] {
	return NewNode(func(nc *NodeContext[struct{}, []byte]) error {
		return pipeReader(nc, r, chunkSize)
	})
}

func pipeReader[T any](nc *NodeContext[T, []byte], r io.Reader, chunkSize int) error {
	for {
		chunk := make([]byte, chunkSize)
		n, err := r.Read(chunk)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		ok := nc.Send(chunk[:n])
		if !ok {
			return nil
		}
	}
}

// Node writing byte chunks into the given write-closer and closing it.
func WriteCloserSink(w io.WriteCloser) *Node[[]byte, struct{}] {
	return NewNode(func(nc *NodeContext[[]byte, struct{}]) (err error) {
		defer func() {
			closeErr := w.Close()
			if err == nil && closeErr != nil {
				err = closeErr
			}
		}()
		for chunk := range nc.Iter() {
			_, err := w.Write(chunk)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Node writing byte chunks into the given writer.
func WriterSink(w io.Writer) *Node[[]byte, struct{}] {
	return NewNode(func(nc *NodeContext[[]byte, struct{}]) error {
		for chunk := range nc.Iter() {
			_, err := w.Write(chunk)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Node reading messages from the given channel.
func ChanSource[T any](ch <-chan T) *Node[struct{}, T] {
	return NewNode(func(nc *NodeContext[struct{}, T]) error {
		for {
			select {
			case msg, more := <-ch:
				if !more {
					return nil
				}
				ok := nc.Send(msg)
				if !ok {
					return nil
				}
			case <-nc.Context().Done():
				return nil
			}
		}
	})
}

// Node writing messages into the given channel.
func ChanSink[T any](ch chan<- T) *Node[T, struct{}] {
	return NewNode(func(nc *NodeContext[T, struct{}]) error {
		for msg := range nc.Iter() {
			select {
			case ch <- msg:
			case <-nc.Context().Done():
				return nil
			}
		}
		return nil
	})
}

func Map[I, O any](h func(I) (O, error)) *Node[I, O] {
	return NewNode(func(nc *NodeContext[I, O]) error {
		for msg := range nc.Iter() {
			res, err := h(msg)
			if err != nil {
				return err
			}
			ok := nc.Send(res)
			if !ok {
				return nil
			}
		}
		return nil
	})
}

func Each[I any](h func(I) error) *Node[I, struct{}] {
	return NewNode(func(nc *NodeContext[I, struct{}]) error {
		for msg := range nc.Iter() {
			err := h(msg)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func Filter[T any](h func(T) (bool, error)) *Node[T, T] {
	return NewNode(func(nc *NodeContext[T, T]) error {
		for msg := range nc.Iter() {
			ok, err := h(msg)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			ok = nc.Send(msg)
			if !ok {
				return nil
			}
		}
		return nil
	})
}

// Given a stream of bytes, split it into lines.
//
// Useful in combination with [CommandSource] to process command stdout line-by-line.
func BytesLinesNode() *Node[[]byte, []byte] {
	return NewNode(func(nc *NodeContext[[]byte, []byte]) error {
		var stdout []byte
		for stdoutChunk := range nc.Iter() {
			stdout = append(stdout, stdoutChunk...)
			for {
				line, rest, found := bytes.Cut(stdout, []byte{'\n'})
				if !found {
					break
				}
				// Not just replace stdout with rest but reallocate it
				// so that the underlying array of the slice doesn't keep growing
				// to the size of the whole stdout.
				stdout = slices.Clone(rest)
				ok := nc.Send(line)
				if !ok {
					return nil
				}
			}
		}
		return nil
	})
}
