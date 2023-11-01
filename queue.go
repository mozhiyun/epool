package epool

import (
	"github.com/eapache/queue/v2"
	"runtime"
	"sync"
)

type Queue[T any] struct {
	sync.Mutex
	popAble *sync.Cond
	buffer  *queue.Queue[T]
	closed  bool
}

// NewQueue 创建队列
func NewQueue[T any]() *Queue[T] {
	e := &Queue[T]{
		buffer: queue.New[T](),
	}
	e.popAble = sync.NewCond(&e.Mutex)
	return e
}

// Push 入队列
func (e *Queue[T]) Push(v T) {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if !e.closed {
		e.buffer.Add(v)
		e.popAble.Signal()
	}
}

// Close 关闭队列
func (e *Queue[T]) Close() {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if !e.closed {
		e.closed = true
		e.popAble.Broadcast() //广播
	}
}

// Pop 取出队列,（阻塞模式）
func (e *Queue[T]) Pop() (v T) {
	c := e.popAble
	buffer := e.buffer

	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	for buffer.Length() == 0 && !e.closed {
		c.Wait()
	}

	if e.closed { //已关闭
		return
	}

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
	}
	return
}

// TryPop 试着取出队列（非阻塞模式）返回ok == false 表示空
func (e *Queue[T]) TryPop() (v T, ok bool) {
	buffer := e.buffer

	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
		ok = true
	} else if e.closed {
		ok = true
	}

	return
}

// Len 获取队列长度
func (e *Queue[T]) Len() int {
	return e.buffer.Length()
}

// Wait 等待队列消费完成
func (e *Queue[T]) Wait() {
	for {
		if e.closed || e.buffer.Length() == 0 {
			break
		}

		runtime.Gosched() //出让时间片
	}
}
