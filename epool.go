package epool

import (
	"sync/atomic"
)

type JobStatistics struct {
	Executing int64 // Total number of jobs executed
	Total     int64
}

type Collector[T any] struct {
	work          chan *Job[T] // receives jobs to send to workers
	end           chan bool    // when receives bool stops workers
	jobS          *JobStatistics
	queue         *Queue[*Job[T]]
	workerChannel chan chan *Job[T]
}

// NewDefault NewDefault(100)
func NewDefault[T any]() *Collector[T] {
	return New[T](100)
}

// New New
func New[T any](workerCount int64) *Collector[T] {
	WorkerChannel := make(chan chan *Job[T])
	var i int64
	var workers []Worker[T]
	input := make(chan *Job[T]) // channel to receive work
	end := make(chan bool)      // channel to spin down workers
	jobFinished := make(chan bool)
	collector := Collector[T]{
		work:          input,
		end:           end,
		jobS:          &JobStatistics{},
		queue:         NewQueue[*Job[T]](),
		workerChannel: WorkerChannel,
	}

	go collector.loopPop()

	for i < workerCount {
		i++
		worker := Worker[T]{
			ID:            i,
			Channel:       make(chan *Job[T]),
			WorkerChannel: WorkerChannel,
			End:           make(chan struct{}),
			jobFinished:   jobFinished,
		}
		worker.Start()
		workers = append(workers, worker) // stores worker
	}
	go func() {
		for {
			select {
			case <-jobFinished: // job finished
				atomic.AddInt64(&collector.jobS.Executing, -1)
			}
		}
	}()

	// start collector
	go func() {
		for {
			select {
			case <-end:
				for _, w := range workers {
					w.Stop() // stop worker
				}
				return
			case job := <-input:
				collector.queue.Push(job)
			}
		}
	}()

	return &collector
}

func (c *Collector[T]) loopPop() {
	for {
		jobObj := c.queue.Pop()
		atomic.AddInt64(&c.jobS.Total, 1)
		worker := <-c.workerChannel // wait for available channel
		atomic.AddInt64(&c.jobS.Executing, 1)
		worker <- jobObj // dispatch work to worker
	}

}

func (c *Collector[T]) GetStatistics() *JobStatistics {
	return c.jobS
}

func (c *Collector[T]) Waiting() int {
	return c.queue.Len()
}

func (c *Collector[T]) AddJob(data T, f JobFunc[T]) {
	c.work <- &Job[T]{Data: data, JobFunc: f}
}

func (c *Collector[T]) Stop() {
	c.end <- true
}
