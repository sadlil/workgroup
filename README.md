# workgroup

[![Go Report Card](https://goreportcard.com/badge/github.com/sadlil/workgroup)](https://goreportcard.com/report/github.com/sadlil/workgroup)
[![Go Reference](https://pkg.go.dev/badge/github.com/sadlil/workgroup.svg)](https://pkg.go.dev/github.com/sadlil/workgroup)
[![Go Coverage](https://github.com/sadlil/workgroup/wiki/coverage.svg)](https://raw.githack.com/wiki/sadlil/workgroup/coverage.html)

**workgroup** is a Go library designed for managing collections of goroutines
that work on subtasks of a common task.

This library is inspired by the [errgroup.Group](https://cs.opensource.google/go/x/sync/+/master:errgroup/)
from the golang.org/x/sync, with additional features and modified behavior for
handling errors and cancellation.

## Features

- **Different Failure Modes**
  - **Collect**: Allows all goroutines to complete, collects all errors, and returns a combined error.
  - **FailFast**: Cancels all remaining goroutines as soon as the first error is encountered and returns that error.
- **Retry**: Support for automated and configurable retries for individual tasks in the group.
- **Concurrency Control**: Configure the maximum number of goroutines that can execute concurrently.

## Acknowledgements

Go Team: This project builds upon the foundation of the errgroup.Group implementation from the golang.org/x/sync package.
The core idea of managing goroutine lifecycle and errors originates from their work.

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

### Workgroup with Retry

```go
ctx, group := workgroup.New(ctx, 
    workgroup.Collect, 
    workgroup.WithRetry(
        retry.Attempts(10), 
        retry.Delay(1*time.Second),
    ))

for i := 0; i < 5; i++ {
    group.Go(ctx, func() error {
        // Simulate a task
        time.Sleep(100 * time.Millisecond)
        return nil
    })
}

group.Wait()
```

Checkout the unit tests for more examples.

## API Documentation

Detailed API documentation can be found at [Godoc](https://pkg.go.dev/github.com/sadlil/workgroup).

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request with improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
