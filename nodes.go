package piper

import (
	"io"
	"os/exec"
)

func CommandSource(cmd *exec.Cmd, chunkSize int) (*Node[struct{}, []byte], error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	node := ReadCloserSource(stdout, chunkSize)
	return node, nil
}

func CommandSink(cmd *exec.Cmd) (*Node[[]byte, struct{}], error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	node := WriteCloserSink(stdin)
	return node, nil
}

func ReadCloserSource(r io.ReadCloser, chunkSize int) *Node[struct{}, []byte] {
	return NewNode(func(nc *NodeContext[struct{}, []byte]) (err error) {
		defer func() {
			closeErr := r.Close()
			if err == nil && closeErr != nil {
				err = closeErr
			}
		}()
		for {
			chunk := make([]byte, chunkSize)
			n, err := r.Read(chunk)
			if err != nil {
				return err
			}
			ok := nc.Send(chunk[:n])
			if !ok {
				return nil
			}
		}
	})
}

func ReaderSource(r io.Reader, chunkSize int) *Node[struct{}, []byte] {
	return NewNode(func(nc *NodeContext[struct{}, []byte]) error {
		for {
			chunk := make([]byte, chunkSize)
			n, err := r.Read(chunk)
			if err != nil {
				return err
			}
			ok := nc.Send(chunk[:n])
			if !ok {
				return nil
			}
		}
	})
}

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
