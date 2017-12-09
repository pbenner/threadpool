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

package threadpool

/* -------------------------------------------------------------------------- */

import "sync"

/* -------------------------------------------------------------------------- */

type job struct {
  f func(ThreadPool, func() error) error
  jobGroup int
}

/* -------------------------------------------------------------------------- */

type waitGroup struct {
  wg    *sync.WaitGroup
  mutex *sync.RWMutex
  cnt    int
}

func newWaitGroup() *waitGroup {
  r := waitGroup{}
  r.wg    = new(sync.WaitGroup)
  r.mutex = new(sync.RWMutex)
  r.cnt   = 0
  return &r
}

func (obj *waitGroup) Value() int {
  obj.mutex.RLock()
  defer obj.mutex.RUnlock()
  return obj.cnt
}

func (obj *waitGroup) Add(i int) {
  obj.mutex.Lock()
  obj.cnt += i
  obj.wg.Add(i)
  obj.mutex.Unlock()
}

func (obj *waitGroup) Done() {
  obj.mutex.Lock()
  obj.cnt -= 1
  obj.wg.Done()
  obj.mutex.Unlock()
}

func (obj *waitGroup) Wait() {
  obj.wg.Wait()
}

/* -------------------------------------------------------------------------- */

type threadPool struct {
  threads  int
  bufsize  int
  channel  chan job
  cntmtx  *sync.RWMutex
  cnt      int
  wgmmtx  *sync.RWMutex
  wgm      map[int]*waitGroup
  errmtx  *sync.RWMutex
  err      map[int]error
}

/* -------------------------------------------------------------------------- */

func (t *threadPool) NewJobGroup() int {
  if t == nil {
    return 0
  }
  t.cntmtx.Lock()
  defer t.cntmtx.Unlock()
  for {
    // increment counter until no wait group is
    // found
    i := t.cnt; t.cnt += 1
    t.wgmmtx.RLock()
    if _, ok := t.wgm[i]; !ok {
      t.wgmmtx.RUnlock()
      return i
    }
    t.wgmmtx.RUnlock()
  }
}

func (t *threadPool) NumberOfThreads() int {
  if t == nil {
    return 1
  } else {
    return t.threads
  }
}

func (t *threadPool) Start() {
  if t == nil {
    return
  }
  t.channel = make(chan job, t.bufsize)
  for i := 1; i < t.threads; i++ {
    go func(i int) {
      for {
        // start computing jobs
        t.worker(i)
      }
    }(i)
  }
}

func (t *threadPool) Stop() {
  if t != nil {
    return
  }
  close(t.channel)
}

/* -------------------------------------------------------------------------- */

func (t *threadPool) setError(jobGroup int, err error) {
  t.errmtx.Lock()
  t.err[jobGroup] = err
  t.errmtx.Unlock()
}

func (t *threadPool) getError(jobGroup int) error {
  t.errmtx.RLock()
  defer t.errmtx.RUnlock()
  if err, ok := t.err[jobGroup]; ok {
    return err
  } else {
    return nil
  }
}

func (t *threadPool) clear(jobGroup int) {
  // clear error
  t.errmtx.Lock()
  delete(t.err, jobGroup)
  t.errmtx.Unlock()
  // clear wait group
  t.wgmmtx.Lock()
  delete(t.wgm, jobGroup)
  t.wgmmtx.Unlock()
}

func (t *threadPool) getWaitGroup(jobGroup int) *waitGroup {
  t.wgmmtx.RLock()
  if wg, ok := t.wgm[jobGroup]; ok {
    t.wgmmtx.RUnlock()
    return wg
  }
  t.wgmmtx.RUnlock()
  // add new wait group
  wg := newWaitGroup()
  t.wgmmtx.Lock()
  t.wgm[jobGroup] = wg
  t.wgmmtx.Unlock()
  return wg
}

func (t *threadPool) worker(i int) {
  for job := range t.channel {
    getError := func() error {
      return t.getError(job.jobGroup)
    }
    if err := job.f(ThreadPool{t, i}, getError); err != nil {
      t.setError(job.jobGroup, err)
    }
  }
}

/* -------------------------------------------------------------------------- */

type ThreadPool struct {
  *threadPool
  // main thread id
  threadId int
}

func (t ThreadPool) GetThreadId() int {
  if t.NumberOfThreads() == 1 {
    return 0
  }
  return t.threadId
}

func (t ThreadPool) AddJob(jobGroup int, f func(pool ThreadPool, erf func() error) error) error {
  if t.NumberOfThreads() == 1 {
    getError := func() error {
      return nil
    }
    if err := f(t, getError); err != nil {
      return err
    }
  } else {
    wg := t.getWaitGroup(jobGroup)
    wg.Add(1)

    g := func(pool ThreadPool, erf func() error) error {
      defer wg.Done()
      return f(pool, erf)
    }
    select {
    case t.channel <- job{g, jobGroup}:
    default:
      // channel buffer is full, execute job here
      getError := func() error {
        return t.getError(jobGroup)
      }
      g(t, getError)
    }
  }
  return nil
}

func (t ThreadPool) AddRangeJob(iFrom, iTo int, jobGroup int, f func(i int, pool ThreadPool, erf func() error) error) error {
  if iFrom >= iTo {
    return nil
  }
  m := t.NumberOfThreads()
  if m > iTo-iFrom {
    m = iTo-iFrom
  }
  n := (iTo-iFrom)/m
  for j := iFrom; j < iTo; j += n {
    iFrom_ := j
    iTo_   := j+n
    if iTo_ > iTo {
      iTo_ = iTo
    }
    if err := t.AddJob(jobGroup, func(pool ThreadPool, erf func() error) error {
      for i := iFrom_; i < iTo_; i++ {
        if err := f(i, pool, erf); err != nil {
          return err
        }
      }
      return nil
    }); err != nil {
      return err
    }
  }
  return nil
}

func (t ThreadPool) Wait(jobGroup int) error {
  if t.NumberOfThreads() == 1 {
    return nil
  }
  t.wgmmtx.RLock()
  if wg, ok := t.wgm[jobGroup]; !ok {
    t.wgmmtx.RUnlock()
    // wait group has not been created, nothing
    // to wait for
    return nil
  } else {
    t.wgmmtx.RUnlock()
    // act as a worker until all jobs of this jobGroup are done
  LOOP:
    for {
      if wg.Value() == 0 {
        break LOOP
      }
      select {
      case job := <- t.channel:
        getError := func() error {
          return t.getError(job.jobGroup)
        }
        if err := job.f(t, getError); err != nil {
          t.setError(job.jobGroup, err)
        }
      default:
        // job channel is empty, wait for all jobs
        // to complete and exit loop
        wg.Wait()
        break LOOP
      }
    }
  }
  // get error message and return
  err := t.getError(jobGroup)
  t.clear(jobGroup)
  return err
}

/* -------------------------------------------------------------------------- */

func NewThreadPool(threads, bufsize int) ThreadPool {
  if threads < 1 {
    panic("invalid number of threads")
  }
  if bufsize < 1 {
    panic("invalid bufsize")
  }
  if threads == 1 {
    return ThreadPool{}
  }
  t := threadPool{}
  t.threads  = threads
  t.bufsize  = bufsize
  t.cntmtx   = new(sync.RWMutex)
  t.cnt      = 0
  t.wgmmtx   = new(sync.RWMutex)
  t.wgm      = make(map[int]*waitGroup)
  t.errmtx   = new(sync.RWMutex)
  t.err      = make(map[int]error)
  // create threads
  t.Start()
  return ThreadPool{&t, 0}
}
