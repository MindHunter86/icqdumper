package app

import (
	"sync"
)

const (
	jobTypeParseChatMess = iota(uint8)
	jobTypeSaveMessage
)

const (
	jobStatusCreated = uint8(iota)
	jobStatusPending
	jobStatusBlocked
	jobStatusFailed
	jobStatusDone
)

type (
	job struct {
		payload     interface{}
		status      uint8
		failedCount uint8
	}
	worker struct {
		pool  chan chan *job
		inbox chan *job

		done chan struct{}
	}
	dispatcher struct {
		queue      chan *job
		pool       chan chan *job
		done       chan struct{}
		workerDone chan struct{}

		workerCapacity int
	}
)

func newDispatcher(queueBuffer, workerCapacity int) *dispatcher {
	return &dispatcher{
		queue:          make(chan *job, queueBuffer),
		pool:           make(chan chan *job, workerCapacity),
		done:           make(chan struct{}, 1),
		workerDone:     make(chan struct{}, 1),
		workerCapacity: workerCapacity,
	}
}

func newWorker(dp *dispatcher) *worker {
	return &worker{
		pool:  dp.pool,
		inbox: make(chan *job, dp.workerCapacity),
		done:  dp.workerDone,
	}
}

func newJob() *job {
	return &job{}
}

func (m *dispatcher) bootstrap(workers int) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers + 1)

	for i := 0; i < workers; i++ {
		go func(wg sync.WaitGroup) {
			newWorker(m).spawn()
			wg.Done()
		}(waitGroup)
	}

	go func(wg sync.WaitGroup) {
		m.dispatch()
		close(m.workerDone)
		wg.Done()
	}(waitGroup)

	waitGroup.Wait()
}

func (m *dispatcher) dispatch() {
	var jbBuf *job
	var nextWorker chan *job

	for {
		select {
		case <-m.done:
			return
		case jbBuf = <-m.queue:
			go func(jb *job) {
				nextWorker = <-m.pool
				nextWorker <- jb
			}(jbBuf)
		}
	}
}

func (m *dispatcher) getQueueChan() chan *job {
	return m.queue
}

func (m *dispatcher) destroy() {
	close(m.done)
}

func (m *worker) spawn() {
	defer close(m.inbox)

	for {
		m.pool <- m.inbox
		select {
		case <-m.done:
		case buf := <-m.inbox:
			m.doJob(buf)
		}
	}

}

func (m *worker) doJob(jb *job) {

}

func (m *job) setStatus(status uint8) (ok bool) {
	m.status, ok = status
	return ok
}

func (m *job) parseChatMessages() (e error) {
	return e
}

func (m *job) saveChatMessage(e error) {
	return e
}
