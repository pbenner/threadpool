/* Copyright (C) 2017 Philipp Benner
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package threadpool

/* -------------------------------------------------------------------------- */

import   "fmt"
import   "testing"
import   "time"

/* -------------------------------------------------------------------------- */

func TestTest1(t *testing.T) {

  n := 10
  p := NewThreadPool(n, 100)
  r := make([]int, n)

  // add jobs
  for i_ := 0; i_ < 100; i_++ {
    i := i_
    p.AddJob(0, func(threadIdx int, erf func() error) error {
      // do nothing if there was an error
      if erf() != nil {
        return nil
      }
      // count the number of jobs this thread
      // finished
      if r[threadIdx] > 3 {
        return fmt.Errorf("error in job %d", i)
      }
      r[threadIdx]++
      return nil
    })
  }
  if err := p.Wait(0); err == nil {
    t.Error("test failed")
  }
}

func TestTest2(t *testing.T) {

  n := 10
  p := NewThreadPool(n, 100)
  r := make([]int, n)

  jobGroup := 0

  // add jobs
  p.AddRangeJob(0, 100, jobGroup, func(i, threadIdx int, erf func() error) error {
    // do nothing if there was an error
    if erf() != nil {
      return nil
    }
    // count the number of jobs this thread
    // finished
    if r[threadIdx] > 3 {
      return fmt.Errorf("error in thread %d", i)
    }
    r[threadIdx]++
    return nil
  })
  if err := p.Wait(jobGroup); err == nil {
    t.Error("test failed:", err)
  }
}

func TestTest3(t *testing.T) {

  n := 1
  p := NewThreadPool(n, 100)
  r := make([]int, n)

  // add jobs
  p.AddRangeJob(0, 100, 0, func(i, threadIdx int, erf func() error) error {
    // do nothing if there was an error
    if erf() != nil {
      return nil
    }
    // count the number of jobs this thread
    // finished
    if r[threadIdx] > 3 {
      return fmt.Errorf("error in thread %d", i)
    }
    r[threadIdx]++
    return nil
  })
  if err := p.Wait(0); err == nil {
    t.Error("test failed")
  }
}

func TestTest4(t *testing.T) {

  n := 10
  m := 5
  p := NewThreadPool(n, 100)
  r := make([]int, m)

  // add jobs
  for i_ := 0; i_ < m; i_++ {
    i := i_
    p.AddJob(0, func(threadIdx int, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddJob(i+1, func(threadIdx int, erf func() error) error {
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
  p := NewThreadPool(n, 100)
  r := make([]int, m)

  // add jobs
  for i_ := 0; i_ < m; i_++ {
    i := i_
    p.AddJob(0, func(threadIdx int, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddJob(i+1, func(threadIdx int, erf func() error) error {
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
    t.Error("test failed: %v", err)
  }
}

/* -------------------------------------------------------------------------- */

// Demonstrate AddJob
func TestExample1(t *testing.T) {

  pool := NewThreadPool(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  for i_, _ := range r {
    i := i_
    pool.AddJob(g, func(threadIdx int, erf func() error) error {
      time.Sleep(10 * time.Millisecond)
      r[i] = threadIdx+1
      return nil
    })
  }
  pool.Wait(g)
  fmt.Println("result:", r)
}

// Demonstrate AddRangeJob
func TestExample2(t *testing.T) {

  pool := NewThreadPool(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  pool.AddRangeJob(0, len(r), g, func(i, threadIdx int, erf func() error) error {
    time.Sleep(10 * time.Millisecond)
    r[i] = threadIdx+1
    return nil
  })
  pool.Wait(g)
  fmt.Println("result:", r)
}

// Demonstrate error handling
func TestExample3(t *testing.T) {

  pool := NewThreadPool(5, 100)

  g := pool.NewJobGroup()
  r := make([]int, 20)

  pool.AddRangeJob(0, len(r), g, func(i, threadIdx int, erf func() error) error {
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
}

// Demonstrate nested job scheduling
func TestExample4(t *testing.T) {

  pool := NewThreadPool(5, 100)

  g0 := pool.NewJobGroup()
  r  := make([][]int, 5)

  pool.AddRangeJob(0, len(r), g0, func(i, threadIdx int, erf func() error) error {
    r[i] = make([]int, 5)

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
}
