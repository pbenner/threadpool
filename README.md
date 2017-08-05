## ThreadPool

Go / Golang thread-pool library that supports nested job queuing.

## Examples

### Example 1: Simple job queuing
```go
  pool := NewThreadPool(5, 100)

  g := pool.NewTaskGroup()
  r := make([]int, 20)

  for i_, _ := range r {
    i := i_
    pool.AddTask(g, func(threadIdx int, erf func() error) error {
      time.Sleep(10 * time.Millisecond)
      r[i] = threadIdx+1
      return nil
    })
  }
  pool.Wait(g)
  fmt.Println("result:", r)
```

### Example 2: Distribute range equally among threads
```go
  pool := NewThreadPool(5, 100)

  g := pool.NewTaskGroup()
  r := make([]int, 20)

  pool.AddRangeTask(0, len(r), g, func(i, threadIdx int, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    r[i] = threadIdx+1
    return nil
  })
  pool.Wait(g)
  fmt.Println("result:", r)
```

### Example 3: Error handling
```go
  pool := NewThreadPool(5, 100)

  g := pool.NewTaskGroup()
  r := make([]int, 20)

  pool.AddRangeTask(0, len(r), g, func(i, threadIdx int, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    if erf() != nil {
      // stop if there was an error
      return nil
    }
    if i == 2 {
      r[i] = -1
      return fmt.Errorf("error in thread %d", threadIdx)
    } else {
      r[i] = threadIdx+1
      return nil
    }
  })
  if err := pool.Wait(g); err != nil {
    fmt.Println(err)
  }
  fmt.Println("result:", r)
```

### Example 4: Nested job queuing
```go
  pool := NewThreadPool(5, 100)

  g0 := pool.NewTaskGroup()
  r  := make([][]int, 5)

  pool.AddRangeTask(0, len(r), g0, func(i, threadIdx int, erf func() error) error {
    r[i] = make([]int, 5)

    gi := pool.NewTaskGroup()

    for j_, _ := range r[i] {
      j := j_
      pool.AddTask(gi, func(threadIdx int, erf func() error) error {
        time.Sleep(10 * time.Millisecond)
        r[i][j] = threadIdx+1
        return nil
      })
    }
    // wait until sub-slice i is filled
    pool.Wait(gi)
    return nil
  })
  // wait until the whole slice is filled
  pool.Wait(g0)
  fmt.Println("result:", r)
```
