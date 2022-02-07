package smartpool

import (
	"runtime"
	"sync/atomic"
)

type (
	worker struct {
		pool *Pool
	}

	Pool struct {
		running uint64
		idle    uint64
		pNum    uint64
		sigChan chan int8
		queue   Queue
	}

	Task interface {
		Execute()
	}
)

func NewPool() *Pool {
	pNum := uint64(runtime.GOMAXPROCS(0))

	p := &Pool{
		sigChan: make(chan int8, pNum),
		queue:   NewSliceQueue(),
		idle:    pNum,
		pNum:    pNum,
	}

	for i := uint64(0); i < pNum; i++ {
		w := &worker{pool: p}
		go w.run()
	}

	return p
}


func (p *Pool) Dispatch(t Task) {
	p.queue.Push(t)

	// TODO when to send a sig?
	if atomic.LoadUint64(&p.running) < p.pNum {
		p.notify()
	}
}

func (p *Pool) notify() {
	select {
	case p.sigChan <- 1:
	default:
	}
}

// check is called when goroutine parks. If the Queue length is 0, do nothing; otherwise, wake up a new worker.
func (p *Pool) check() {
	atomic.AddUint64(&p.running, -1)
	if p.queue.RoughLen() > 0 {
		if atomic.CompareAndSwapUint64(&p.idle,0,1) {
			w := &worker{pool: p}
			go w.run()
		}
		p.notify()
	}
}

func (w *worker) run() {
	var t Task
	for range w.pool.sigChan {
		atomic.AddUint64(&w.pool.idle, -1)
		atomic.AddUint64(&w.pool.running, 1)

		for t = w.pool.queue.Pop(); t != nil; {
			t.Execute()
		}


		if atomic.LoadUint64(&w.pool.idle) < w.pool.pNum {
			atomic.AddUint64(&w.pool.idle, 1)
			continue
		}

		// retire from work
		atomic.AddUint64(&w.pool.idle, -1)
		atomic.AddUint64(&w.pool.running, -1)
		break
	}
}
