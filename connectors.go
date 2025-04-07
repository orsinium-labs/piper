package piper

import "context"

// Connect two nodes together.
func Connect[T, X, Y any](
	n1 *Node[X, T],
	n2 *Node[T, Y],
) {
	ch := make(chan T)
	ConnectChan(n1, n2, ch)
}

// Connect 3 nodes together.
func Connect3[A, B, C, D any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
) {
	Connect(n1, n2)
	Connect(n2, n3)
}

// Connect 4 nodes together.
func Connect4[A, B, C, D, E any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
}

// Connect 5 nodes together.
func Connect5[A, B, C, D, E, F any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
	Connect(n4, n5)
}

// Connect 6 nodes together.
func Connect6[A, B, C, D, E, F, G any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
	Connect(n4, n5)
	Connect(n5, n6)
}

// Connect two nodes together using the provided channel.
func ConnectChan[T, X, Y any](
	n1 *Node[X, T],
	n2 *Node[T, Y],
	ch chan T,
) {
	n1.context.out.ch = ch
	n2.context.in.ch = ch

	done := make(chan struct{})
	n1.context.out.done = done
	n2.context.in.done = done
}

// Connect and run the given 2 nodes.
func Pipe2[T, X, Y any](
	ctx context.Context,
	n1 *Node[X, T],
	n2 *Node[T, Y],
) Errors {
	Connect(n1, n2)
	return Run(ctx, n1, n2)
}

// Connect and run the given 3 nodes.
func Pipe3[A, B, C, D any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
) Errors {
	Connect3(n1, n2, n3)
	return Run(ctx, n1, n2, n3)
}

// Connect and run the given 4 nodes.
func Pipe4[A, B, C, D, E any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
) Errors {
	Connect4(n1, n2, n3, n4)
	return Run(ctx, n1, n2, n3, n4)
}

// Connect and run the given 5 nodes.
func Pipe5[A, B, C, D, E, F any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
) Errors {
	Connect5(n1, n2, n3, n4, n5)
	return Run(ctx, n1, n2, n3, n4, n5)
}

// Connect and run the given 6 nodes.
func Pipe6[A, B, C, D, E, F, G any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
) Errors {
	Connect6(n1, n2, n3, n4, n5, n6)
	return Run(ctx, n1, n2, n3, n4, n5, n6)
}
