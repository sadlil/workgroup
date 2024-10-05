# workgroup

[![Go Report Card](https://goreportcard.com/badge/github.com/sadlil/workgroup)](https://goreportcard.com/report/github.com/sadlil/workgroup)
[![Sourcegraph](https://sourcegraph.com/github.com/sadlil/workgroup/-/badge.svg)](https://sourcegraph.com/github.com/sadlil/workgroup?badge)
[![Go Reference](https://pkg.go.dev/badge/github.com/sadlil/workgroup.svg)](https://pkg.go.dev/github.com/sadlil/workgroup)


workgroup is a Go library designed for managing collections of goroutines that work on subtasks of a common task. It offers enhanced error propagation, context cancellation signals, and goroutine synchronization mechanisms.

This library is a fork of the errgroup.Group from x/sync, with additional features and modified behavior for handling errors and cancellation.

## Acknowledgements

Go Team: This project builds upon the foundation of the errgroup.Group implementation from the golang.org/x/sync package. The core idea of managing goroutine lifecycle and errors originates from their work.

## Features

Failure Modes:

- Collect: Allows all goroutines to complete, collects all errors, and returns a combined error.
- FailFast: Cancels all remaining goroutines as soon as the first error is encountered and returns that error.
- Concurrency Control: Configure the maximum number of goroutines that can execute concurrently.

## Installation

To get the workgroup package, run:

```bash
go get github.com/sadlil/workgroup
```

## Examples

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/sadlil/workgroup"
)

func main() {
    ctx := context.Background()
    
    // Create a new workgroup with Collect failure mode
    ctx, g := workgroup.New(ctx, workgroup.Collect)

    g.Go(ctx, func() error {
        // Simulate a task
        time.Sleep(100 * time.Millisecond)
        return errors.New("task 1 failed")
    })

    g.Go(ctx, func() error {
        time.Sleep(200 * time.Millisecond)
        return errors.New("task 2 failed")
    })

    // Wait for all goroutines to complete
    if err := g.Wait(); err != nil {
        fmt.Printf("Workgroup completed with error: %v\n", err)
    }
}
```

### Example with FailFast Mode

In FailFast mode, the first error cancels all running goroutines and returns that error:

```go
ctx, g := workgroup.New(ctx, workgroup.FailFast)

g.Go(ctx, func() error {
    time.Sleep(100 * time.Millisecond)
    return errors.New("this error cancels all other tasks")
})

g.Go(ctx, func() error {
    // This will be canceled before it completes
    time.Sleep(200 * time.Millisecond)
    return nil
})

err := g.Wait()
fmt.Printf("Error encountered: %v\n", err)
```

### Workgroup with Limit

```go
ctx, g := workgroup.New(ctx, workgroup.Collect, workgroup.WithLimit(2))

for i := 0; i < 5; i++ {
    g.Go(ctx, func() error {
        // Simulate a task
        time.Sleep(100 * time.Millisecond)
        return nil
    })
}

g.Wait()
```

Checkout the unit tests for more examples.

## API Documentation

Detailed API documentation can be found at [Godoc](https://pkg.go.dev/github.com/sadlil/workgroup).

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request with improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
