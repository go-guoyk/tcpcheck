package main

import "time"

type StopWatch struct {
	t time.Time
}

func (s *StopWatch) Reset() {
	s.t = time.Now()
}

func (s *StopWatch) Stop() int64 {
	return int64(time.Now().Sub(s.t) / time.Millisecond)
}

func NewStopWatch() *StopWatch {
	return &StopWatch{t: time.Now()}
}
