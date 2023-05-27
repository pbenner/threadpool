/* Copyright (C) 2017-2023 Philipp Benner
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package threadpool

/* -------------------------------------------------------------------------- */

import "fmt"
import "testing"
import "time"

/* -------------------------------------------------------------------------- */

func TestTest1(t *testing.T) {

  n := 10
  p := New(n, 100)
  r := make([]int, n)

  // add jobs
  for i_ := 0; i_ < 100; i_++ {
    i := i_
    p.AddJob(0, func(p ThreadPool, erf func() error) error {
      // do nothing if there was an error
      if erf() != nil {
        return nil
      }
      // count the number of jobs this thread
      // finished
      if r[p.GetThreadId()] > 3 {
        return fmt.Errorf("error in job %d", i)
      }
      r[p.GetThreadId()]++
      return nil
    })
  }
  if err := p.Wait(0); err == nil {
    t.Error("test failed")
  }
}

func TestTest2(t *testing.T) {

  n := 10
  p := New(n, 100)
  r := make([]int, n)

  jobGroup := 0

  // add jobs
  p.AddRangeJob(0, 100, jobGroup, func(i int, p ThreadPool, erf func() error) error {
    // do nothing if there was an error
    if erf() != nil {
      return nil
    }
    // count the number of jobs this thread
    // finished
    if r[p.GetThreadId()] > 3 {
      return fmt.Errorf("error in thread %d", i)
    }
    r[p.GetThreadId()]++
    return nil
  })
  if err := p.Wait(jobGroup); err == nil {
    t.Error("test failed:", err)
  }
}

func TestTest3(t *testing.T) {

  n := 1
  p := New(n, 100)
  r := make([]int, n)

  // add jobs
  if err := p.AddRangeJob(0, 100, 0, func(i int, p ThreadPool, erf func() error) error {
    // do nothing if there was an error
    if erf() != nil {
      return nil
    }
    // count the number of jobs this thread
    // finished
    if r[p.GetThreadId()] > 3 {
      return fmt.Errorf("error in thread %d", i)
    }
    r[p.GetThreadId()]++
    return nil
  }); err == nil {
    t.Error("test failed")
  }
  if err := p.Wait(0); err != nil {
    t.Error("test failed")
  }
}

func TestTest4(t *testing.T) {

  n := 10
  m := 5
  p := New(n, 100)
  r := make([]int, m)

  // add jobs
  for i_ := 0; i_ < m; i_++ {
    i := i_
    p.AddJob(0, func(p ThreadPool, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddJob(i+1, func(p ThreadPool, erf func() error) error {
          r[i]++
          return nil
        })
      }
      if err := p.Wait(i+1); err != nil {
        t.Error("test failed")
      }
      return nil
    })
  }
  if err := p.Wait(0); err != nil {
    t.Error("test failed")
  }
}

func TestTest5(t *testing.T) {

  n := 1
  m := 5
  p := New(n, 100)
  r := make([]int, m)

  // add jobs
  for i_ := 0; i_ < m; i_++ {
    i := i_
    p.AddJob(0, func(p ThreadPool, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddJob(i+1, func(p ThreadPool, erf func() error) error {
          r[i]++
          return nil
        })
      }
      if err := p.Wait(i+1); err != nil {
        t.Errorf("test failed: %v", err)
      }
      return nil
    })
  }
  if err := p.Wait(0); err != nil {
    t.Errorf("test failed: %v", err)
  }
}

/* -------------------------------------------------------------------------- */

// Demonstrate AddJob
func TestExample1(t *testing.T) {
  // create a new thread pool with 5 working threads and
  // a queue buffer of 100 (in addition to this thread, 4
  // more threads will be launched that start reading
  // from the job queue)
  pool := New(5, 100)

  // jobs are always grouped, get a new group index
  g := pool.NewJobGroup()
  // slice carrying the results
  r := make([]int, 20)

  // add jobs to the thread pool, where the i'th job sets
  // r[i] to the thread index
  for i_, _ := range r {
    i := i_
    pool.AddJob(g, func(pool ThreadPool, erf func() error) error {
      time.Sleep(10 * time.Millisecond)
      r[i] = pool.GetThreadId()+1
      return nil
    })
  }
  // wait until all jobs in group g are done, meanwhile, this thread
  // is also used as a worker
  pool.Wait(g)
  fmt.Println("result:", r)
}

// Demonstrate AddRangeJob
func TestExample2(t *testing.T) {

  pool := New(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  // instead of creating len(r) jobs, this method splits
  // r into #threads pieces and adds one job for each piece
  // to increase efficiency
  pool.AddRangeJob(0, len(r), g, func(i int, pool ThreadPool, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    r[i] = pool.GetThreadId()+1
    return nil
  })
  pool.Wait(g)
  fmt.Println("result:", r)
}

// Demonstrate error handling
func TestExample3(t *testing.T) {

  pool := New(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  if err := pool.AddRangeJob(0, len(r), g, func(i int, pool ThreadPool, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    // stop if there was an error in one of the
    // previous jobs
    if erf() != nil {
      // stop if there was an error
      return nil
    }
    if i == 2 {
      r[i] = -1
      return fmt.Errorf("error in thread %d", pool.GetThreadId())
    } else {
      r[i] = pool.GetThreadId()+1
      return nil
    }
  }); err != nil {
    fmt.Println(err)
  }
  if err := pool.Wait(g); err != nil {
    fmt.Println(err)
  }
  fmt.Println("result:", r)
}

// Demonstrate nested job scheduling
func TestExample4(t *testing.T) {

  pool := New(5, 100)

  g0 := pool.NewJobGroup()
  r  := make([][]int, 5)

  pool.AddRangeJob(0, len(r), g0, func(i int, pool ThreadPool, erf func() error) error {
    r[i] = make([]int, 5)

    // get a new job group for filling the i'th sub-slice, which allows
    // us to wait until the sub-slice is filled
    gi := pool.NewJobGroup()

    for j_, _ := range r[i] {
      j := j_
      pool.AddJob(gi, func(pool ThreadPool, erf func() error) error {
        time.Sleep(10 * time.Millisecond)
        r[i][j] = pool.GetThreadId()+1
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
}
