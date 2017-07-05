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
  for i := 0; i < 100; i++ {
    p.AddTask(func(i int, erf func() error) error {
      // do nothing if there was an error
      if erf() != nil {
        return nil
      }
      // count the number of jobs this thread
      // finished
      if r[i] > 3 {
        return fmt.Errorf("error in thread %d", i)
      }
      r[i]++
      return nil
    })
  }
  if err := p.Wait(); err == nil {
    t.Error("test failed")
  }
}
