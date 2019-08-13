package scheduledinc

import (
	"sync"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/robfig/cron"
)

type Scheduler struct {
	cron      *cron.Cron
	cronReset chan struct{}

	// cron does not have a say to remove an item
	items     map[string]ScheduledIncident
	itemsLock sync.RWMutex

	logTag string
	logger boshlog.Logger
}

func NewScheduler(logger boshlog.Logger) *Scheduler {
	return &Scheduler{
		cronReset: make(chan struct{}),

		items: map[string]ScheduledIncident{},

		logTag: "Scheduler",
		logger: logger,
	}
}

func (s *Scheduler) Run() {
	for _ = range s.cronReset {
		s.performCronReset()
	}
}

func (s Scheduler) ScheduledIncidentWasCreated(si ScheduledIncident) {
	s.itemsLock.Lock()
	s.items[si.ID] = si
	s.itemsLock.Unlock()

	s.signalCronReset()
}

func (s Scheduler) ScheduledIncidentWasDeleted(si ScheduledIncident) {
	s.itemsLock.Lock()
	delete(s.items, si.ID)
	s.itemsLock.Unlock()

	s.signalCronReset()
}

func (s Scheduler) signalCronReset() {
	select {
	case s.cronReset <- struct{}{}:
		// signalled
	default:
		// ignored since already resetting
	}
}

func (s *Scheduler) performCronReset() {
	if s.cron != nil {
		s.cron.Stop()
	}

	s.cron = cron.New()

	s.itemsLock.Lock()

	for i, _ := range s.items {
		si := s.items[i]
		s.cron.AddFunc(si.Schedule, func() {
			err := si.Execute()
			if err != nil {
				s.logger.Error(s.logTag, "Failed to queue up scheduled incident: %s", err.Error())
			}
		})
	}

	s.itemsLock.Unlock()

	s.cron.Start()
}
