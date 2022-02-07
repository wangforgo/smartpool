package smartpool

import "sync"

type Queue interface {
	Push(task Task)
	Pop() Task
	Len() int      // accurate length of the Queue.
	RoughLen() int // rough length, for fast check.
}

type sliceQueue struct {
	data []Task
	n    int
	sync.Mutex
}

func NewSliceQueue() Queue {
	return &sliceQueue{
		data: make([]Task, 100),
	}
}

func (s *sliceQueue) Push(task Task) {
	s.Lock()
	s.data = append(s.data, task)
	s.n++
	s.Unlock()
}

func (s *sliceQueue) Pop() Task {
	var t Task
	s.Lock()
	t = s.data[0]
	s.data = s.data[1:]
	s.n++
	s.Unlock()
	return t
}

func (s *sliceQueue) Len() int {
	var n int
	s.Lock()
	n = s.n
	s.Unlock()
	return n
}

func (s *sliceQueue) RoughLen() int {
	return s.n
}
