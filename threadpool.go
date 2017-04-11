/* Copyright (C) 2016 Philipp Benner
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

package utility

/* -------------------------------------------------------------------------- */

import "sync"

/* -------------------------------------------------------------------------- */

type ThreadPool struct {
  threads int
  bufsize int
  channel chan func(int)
  wg      sync.WaitGroup
}

func NewThreadPool(threads, bufsize int) *ThreadPool {
  t := ThreadPool{}
  t.threads = threads
  t.bufsize = bufsize
  t.Launch()
  return &t
}

func (t *ThreadPool) AddTask(task func(i int)) {
  t.channel <- task
}

func (t *ThreadPool) Wait() {
  close(t.channel)
  t.wg.Wait()
  t.Launch()
}

func (t *ThreadPool) Launch() {
  t.channel = make(chan func(int), t.bufsize)
  t.wg.Add(t.threads)
  for i := 0; i < t.threads; i++ {
    go func(i int) {
      defer t.wg.Done()
      for task := range t.channel {
        task(i)
      }
    }(i)
  }
}

func (t *ThreadPool) NumberOfThreads() int {
  return t.threads
}
