## ThreadPool

Go / Golang thread-pool library that supports nested job queuing.

## Examples

### Example 1: Simple job queuing
```go
  // create a new thread pool with 5 working threads and
  // a queue buffer of 100 (in addition to this thread, 4
  // additional threads will be launched that start reading
  // from the job queue)
  pool := NewThreadPool(5, 100)

  // jobs are always grouped, get a new group index
  g := pool.NewJobGroup()
  // slice carrying the results
  r := make([]int, 20)

  // add jobs to the thread pool, where the i'th job sets
  // r[i] to the thread index
  for i_, _ := range r {
    i := i_
    pool.AddJob(g, func(threadIdx int, erf func() error) error {
      time.Sleep(10 * time.Millisecond)
      r[i] = threadIdx+1
      return nil
    })
  }
  // wait until all jobs in group g are done, meanwhile, this thread
  // is also used as a worker
  pool.Wait(g)
  fmt.Println("result:", r)
```

### Example 2: Distribute range equally among threads
```go
  pool := NewThreadPool(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  // instead of creating len(r) jobs, this method splits
  // r into #threads pieces and adds one job for each piece
  // to increase efficiency
  pool.AddRangeJob(0, len(r), g, func(i, threadIdx int, erf func() error) error {
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

  g := pool.NewJobGroup()
  r := make([]int, 20)

  pool.AddRangeJob(0, len(r), g, func(i, threadIdx int, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    // stop if there was an error in one of the
    // previous jobs
    if erf() != nil {
      return nil
    }
    // return an error at position 2
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

  g0 := pool.NewJobGroup()
  r  := make([][]int, 5)

  pool.AddRangeJob(0, len(r), g0, func(i, threadIdx int, erf func() error) error {
    r[i] = make([]int, 5)

    // get a new job group for filling the i'th sub-slice, which allows
    // us to wait until the sub-slice is filled
    gi := pool.NewJobGroup()

    for j_, _ := range r[i] {
      j := j_
      pool.AddJob(gi, func(threadIdx int, erf func() error) error {
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
