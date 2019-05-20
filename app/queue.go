package app

import (
	"sync"
)

const (
	jobActParseChatMessages = uint8(iota)
	jobActSaveChatMessage
	jobActCustomFunc
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
		payload     []interface{}
		payloadFunc func([]interface{}) error
		status      uint8
		action      uint8
		failedCount uint8
	}
	jobError struct {
		e   error
		job *job
	}
	worker struct {
		pool  chan chan *job
		inbox chan *job

		done   chan struct{}
		errors chan *jobError
	}
	dispatcher struct {
		queue      chan *job
		pool       chan chan *job
		done       chan struct{}
		workerDone chan struct{}
		errorPipe  chan *jobError

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
		errorPipe:      make(chan *jobError),
	}
}

func newWorker(dp *dispatcher) *worker {
	return &worker{
		pool:   dp.pool,
		inbox:  make(chan *job, dp.workerCapacity),
		done:   dp.workerDone,
		errors: dp.errorPipe,
	}
}

func newJob() *job {
	return &job{
		status: jobStatusCreated,
	}
}

func (m *dispatcher) bootstrap(workers int) (e error) {
	gLogger.Debug().Msg("Starting worker spawning...")

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

	gLogger.Info().Int("workers count", workers).Msg("Workers has been spawned successfully")
	waitGroup.Wait()

	return
}

func (m *dispatcher) dispatch() {
	var jbBuf *job

	gLogger.Info().Msg("Dispatching has been successfully started")

	for {
		select {
		case <-m.done:
			return
		case jbBuf = <-m.queue:
			go func(jb *job) {
				nextWorker := <-m.pool
				nextWorker <- jb
			}(jbBuf)
		case jbErr := <-m.errorPipe:
			if jbErr.job.failedCount != 3 {
				gLogger.Info().Msg("Trying to restart failed job...")
				m.queue <- jbErr.job
			} else {
				gLogger.Error().Msg("Could not restart failed job! Fails count is more or equal 3!")
			}
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
			buf.setStatus(jobStatusPending)
			m.doJob(buf)
		}
	}
}

func (m *worker) doJob(jb *job) {

	switch jb.action {
	case jobActParseChatMessages:
	case jobActSaveChatMessage:
	case jobActCustomFunc:
		if jb.payloadFunc != nil {
			if e := jb.payloadFunc(jb.payload); e != nil {
				m.errors <- jb.newError(e)
			}
		} else {
			gLogger.Warn().Msg("Job has undefined action")
		}
	}
}

func (m *job) setStatus(status uint8) {
	if status == jobStatusFailed {
		m.failedCount++
	}

	m.status = status
}

func (m *job) newError(e error) *jobError {
	m.setStatus(jobStatusFailed)
	gLogger.Warn().Err(e).Uint8("failed tries", m.failedCount).Msg("Could not exec job. Job state now is Failed")
	return &jobError{
		e:   e,
		job: m,
	}
}

func (m *job) parseChatMessages() (e error) {
	return
}

func (m *job) saveChatMessage(e error) {
	return
}
