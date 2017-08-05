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
//import   "time"

/* -------------------------------------------------------------------------- */

func TestTest1(t *testing.T) {

  n := 10
  p := NewThreadPool(n, 100)
  r := make([]int, n)

  // add jobs
  for i_ := 0; i_ < 100; i_++ {
    i := i_
    p.AddTask(0, func(threadIdx int, erf func() error) error {
      // do nothing if there was an error
      if erf() != nil {
        return nil
      }
      // count the number of jobs this thread
      // finished
      if r[threadIdx] > 3 {
        return fmt.Errorf("error in task %d", i)
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

  taskGroup := 0

  // add jobs
  p.AddRangeTask(0, 100, taskGroup, func(i, threadIdx int, erf func() error) error {
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
  if err := p.Wait(taskGroup); err == nil {
    t.Error("test failed:", err)
  }
}

func TestTest3(t *testing.T) {

  n := 1
  p := NewThreadPool(n, 100)
  r := make([]int, n)

  // add jobs
  p.AddRangeTask(0, 100, 0, func(i, threadIdx int, erf func() error) error {
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
    p.AddTask(0, func(threadIdx int, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddTask(i+1, func(threadIdx int, erf func() error) error {
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
    p.AddTask(0, func(threadIdx int, erf func() error) error {
      for j := 0; j < m; j++ {
        p.AddTask(i+1, func(threadIdx int, erf func() error) error {
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
