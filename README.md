# piper

[ [📄 docs](https://pkg.go.dev/github.com/orsinium-labs/piper) ] [ [🐙 github](https://github.com/orsinium-labs/piper) ]

Go package for building type-safe, composable, and cancellable pipelines.

Features:

* Type safety
* Concurrency
* Error propagation
* Cancellation

## How it works

You create a pipeline: a series of nodes connected with each other. Internally, each node is a goroutine and each connection is a channel. Each node implementation accepts a `NodeContext` which has 3 methods:

* `Recv` to read an incoming message.
* `Send` to write an outgoing message.
* `Error` to emit an error without stopping the node.

Internally, the package takes care of many concurrency corner-cases:

* If the node exits, all incoming nodes will also exit (their `Send` will return `false`): if there is no node to handle produced values, we should stop producing the values.
* If the node exits, the channel for the outgoing nodes will be closed. After the outgoing node consumes all values, it will also exit (`Recv` will return `false`).
* If the context is cancelled, all nodes exit (all `Recv` and `Send` will return `false`).

## Installation

```bash
go get github.com/orsinium-labs/piper
```

## Usage

Make a node producing values ("source"):

```go
numbers := piper.NewNode(func(nc *piper.NodeContext[struct{}, int]) error {
    for i := 1; i < 100; i++ {
        ok := nc.Send(i)
        if !ok {
          return nil
        }
    }
    return nil
})
```

Make a node processing the values:

```go
doubler := piper.NewNode(func(nc *piper.NodeContext[int, int]) error {
    for n := range nc.Iter() {
        ok := nc.Send(n * 2)
        if !ok {
            break
        }
    }
    return nil
})
```

The same node can be simplified using `Map`:

```go
doubler := piper.Map(func(n int) (int, error) {
    return n * 2, nil
})
```

Make a node collecting all the values together ("reduce sink"):

```go
sum := 0
summer := piper.NewNode(func(nc *piper.NodeContext[int, struct{}]) error {
    for n := range nc.Iter() {
        sum += n
    }
    return nil
})
```

The same node can be simplified using `Each`:

```go
sum := 0
summer := piper.Each(func(n int) error {
    sum += n
    return nil
})
```

Connect all nodes together, run the pipeline in background, and wait for it to finish:

```go
piper.Connect(numbers, doubler)
piper.Connect(doubler, summer)
err := piper.Wait(piper.Run(ctx, numbers, doubler, summer))
```

Or the same using a `Pipe` shortcut:

```go
err := piper.Wait(piper.Pipe3(ctx, numbers, doubler, summer))
```
