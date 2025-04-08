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

// Connect 3 nodes together sequentially.
func Connect3[A, B, C, D any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
) {
	Connect(n1, n2)
	Connect(n2, n3)
}

// Connect 4 nodes together sequentially.
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

// Connect 5 nodes together sequentially.
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

// Connect 6 nodes together sequentially.
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

// Connect 7 nodes together sequentially.
func Connect7[A, B, C, D, E, F, G, H any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
	Connect(n4, n5)
	Connect(n5, n6)
	Connect(n6, n7)
}

// Connect 8 nodes together sequentially.
func Connect8[A, B, C, D, E, F, G, H, J any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
	n8 *Node[H, J],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
	Connect(n4, n5)
	Connect(n5, n6)
	Connect(n6, n7)
	Connect(n7, n8)
}

// Connect 9 nodes together sequentially.
func Connect9[A, B, C, D, E, F, G, H, J, K any](
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
	n8 *Node[H, J],
	n9 *Node[J, K],
) {
	Connect(n1, n2)
	Connect(n2, n3)
	Connect(n3, n4)
	Connect(n4, n5)
	Connect(n5, n6)
	Connect(n6, n7)
	Connect(n7, n8)
	Connect(n8, n9)
}

// Connect two nodes together using the provided channel.
func ConnectChan[T, X, Y any](
	n1 *Node[X, T],
	n2 *Node[T, Y],
	ch chan T,
) {
	if n1.context.out.ch != nil {
		panic("one-to-many connection is not implemented yet")
	}
	if n2.context.in.ch != nil {
		panic("many-to-one connection is not implemented yet")
	}
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

// Connect and run the given 7 nodes.
func Pipe7[A, B, C, D, E, F, G, H any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
) Errors {
	Connect7(n1, n2, n3, n4, n5, n6, n7)
	return Run(ctx, n1, n2, n3, n4, n5, n6, n7)
}

// Connect and run the given 8 nodes.
func Pipe8[A, B, C, D, E, F, G, H, J any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
	n8 *Node[H, J],
) Errors {
	Connect8(n1, n2, n3, n4, n5, n6, n7, n8)
	return Run(ctx, n1, n2, n3, n4, n5, n6, n7, n8)
}

// Connect and run the given 9 nodes.
func Pipe9[A, B, C, D, E, F, G, H, J, K any](
	ctx context.Context,
	n1 *Node[A, B],
	n2 *Node[B, C],
	n3 *Node[C, D],
	n4 *Node[D, E],
	n5 *Node[E, F],
	n6 *Node[F, G],
	n7 *Node[G, H],
	n8 *Node[H, J],
	n9 *Node[J, K],
) Errors {
	Connect9(n1, n2, n3, n4, n5, n6, n7, n8, n9)
	return Run(ctx, n1, n2, n3, n4, n5, n6, n7, n8, n9)
}
