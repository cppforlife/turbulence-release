package tasks

import (
	"sync"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type repo struct {
	inboxes     map[string]agentInbox
	inboxesLock sync.RWMutex

	tasks     map[string]TaskReq
	taskChs   map[string]chan struct{}
	tasksLock sync.RWMutex

	logTag string
	logger boshlog.Logger
}

type agentInbox struct {
	consumed chan struct{}
	tasks    []Task
}

func NewRepo(logger boshlog.Logger) Repo {
	return &repo{
		inboxes: map[string]agentInbox{},

		tasks:   map[string]TaskReq{},
		taskChs: map[string]chan struct{}{},

		logTag: "tasks.repo",
		logger: logger,
	}
}

func (r *repo) QueueAndWait(agentID string, tasks []Task) error {
	// Set up wait channels for tasks
	r.tasksLock.Lock()

	for _, task := range tasks {
		r.taskChs[task.ID] = make(chan struct{})
	}

	r.tasksLock.Unlock()

	// Set up agent inbox
	r.inboxesLock.Lock()

	rec, found := r.inboxes[agentID]
	if !found {
		rec = agentInbox{consumed: make(chan struct{}, 0)}
	}

	rec.tasks = append(rec.tasks, tasks...)

	r.inboxes[agentID] = rec

	// Local ref before unlocking
	consumed := rec.consumed

	// Unlock before blocking
	r.inboxesLock.Unlock()

	select {
	case <-consumed:
		r.logger.Debug(r.logTag, "Finished waiting since agent '%s' consumed tasks", agentID)
		return nil
	case <-time.After(30 * time.Second):
		r.logger.Error(r.logTag, "Timed out waiting for agent '%s' to consume tasks", agentID)

		// Clean up agent inbox
		r.inboxesLock.Lock()

		delete(r.inboxes, agentID)

		r.inboxesLock.Unlock()

		// Clean up task inboxes
		r.tasksLock.Lock()

		for _, task := range tasks {
			delete(r.taskChs, task.ID)
		}

		r.tasksLock.Unlock()

		return bosherr.Errorf("Timed out waiting for agent '%s' to consume tasks", agentID)
	}
}

func (r *repo) Consume(agentID string) ([]Task, error) {
	if len(agentID) == 0 {
		return nil, bosherr.Error("Must provide non-empty agent ID")
	}

	r.inboxesLock.Lock()
	defer r.inboxesLock.Unlock()

	rec, found := r.inboxes[agentID]
	if found {
		// Unblock all waiting clients
		close(rec.consumed)

		// Reset agent inbox
		delete(r.inboxes, agentID)

		r.logger.Debug(r.logTag, "Consumed tasks for agent '%s'", agentID)
	} else {
		// Too noisy...
		// r.logger.Debug(r.logTag, "Consumed no tasks for agent '%s'", agentID)
	}

	return rec.tasks, nil
}

func (r *repo) Wait(taskID string) (TaskReq, error) {
	if len(taskID) == 0 {
		return TaskReq{}, bosherr.Error("Must provide non-empty task ID")
	}

	r.tasksLock.Lock()

	ch, found := r.taskChs[taskID]
	if !found {
		return TaskReq{}, bosherr.Error("Waiting must happen after queueing")
	}

	r.tasksLock.Unlock()

	<-ch

	// Fetch saved task
	r.tasksLock.Lock()
	defer r.tasksLock.Unlock()

	return r.tasks[taskID], nil
}

func (r *repo) Update(taskID string, taskReq TaskReq) error {
	if len(taskID) == 0 {
		return bosherr.Error("Must provide non-empty task ID")
	}

	r.tasksLock.Lock()
	defer r.tasksLock.Unlock()

	// Save task before closing channel
	r.tasks[taskID] = taskReq

	if ch, found := r.taskChs[taskID]; found {
		// Unblock all waiting clients
		close(ch)

		// Reset task
		delete(r.taskChs, taskID)

		r.logger.Debug(r.logTag, "Updated task '%s'", taskID)
	} else {
		r.logger.Debug(r.logTag, "Did not wait for task '%s'", taskID)
	}

	return nil
}
