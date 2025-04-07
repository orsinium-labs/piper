package piper

import "io"

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
