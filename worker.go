package epool

import (
	"log"
)

type JobFunc[T any] func(id int64, data T)

type Job[T any] struct {
	Data    T
	JobFunc JobFunc[T]
}

type Worker[T any] struct {
	ID            int64
	WorkerChannel chan chan *Job[T] // used to communicate between epool and workers
	Channel       chan *Job[T]
	End           chan struct{}
	jobFinished   chan bool
}

// Start  worker
func (w *Worker[T]) Start() {
	go func() {
		for {
			w.WorkerChannel <- w.Channel // when the worker is available place channel in queue
			select {
			case job := <-w.Channel: // worker has received job
				if job != nil {
					job.JobFunc(w.ID, job.Data) // do work
					w.jobFinished <- true
				}

			case <-w.End:
				return
			}
		}
	}()
}

// Stop  end worker
func (w *Worker[T]) Stop() {
	log.Printf("worker [%d] is stopping", w.ID)
	w.End <- struct{}{}
}
