package healthcheck

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrHealthcheckerNotStarted     = errors.New("HealthChecker not started yet.")
	ErrHealthcheckerAlreadyStarted = errors.New("HealthChecker already started.")
)

type HealthIndicator interface {
	Name() string
	IsHealthy() bool
}

type HealthCheckerHook func(map[string]bool)

type HealthChecker interface {
	Start() error
	Stop() error
	RegisterIndicator(indicator HealthIndicator)
	UnregisterIndicator(indicator HealthIndicator)
	AddHook(name string, hook HealthCheckerHook)
}

type AsyncHealthChecker struct {
	period        time.Duration
	indicatorLock *sync.Mutex
	hookLock      *sync.Mutex
	indicators    map[string]HealthIndicator
	hooks         map[string]HealthCheckerHook
	stopChan      chan struct{}
}

func New(period time.Duration) HealthChecker {
	checker := &AsyncHealthChecker{
		period:        period,
		indicatorLock: &sync.Mutex{},
		hookLock:      &sync.Mutex{},
		indicators:    make(map[string]HealthIndicator),
		hooks:         make(map[string]HealthCheckerHook),
	}

	return checker
}

func (checker *AsyncHealthChecker) Start() error {
	if checker.stopChan != nil {
		return ErrHealthcheckerAlreadyStarted
	}

	checker.stopChan = make(chan struct{}, 0)
	go func(ch *AsyncHealthChecker) {
		for {
			time.Sleep(ch.period)
			select {
			case <-ch.stopChan:
				close(checker.stopChan)
				return
			default:
				res := make(map[string]bool)
				ch.indicatorLock.Lock()
				for _, indicator := range ch.indicators {
					res[indicator.Name()] = indicator.IsHealthy()
				}
				ch.indicatorLock.Unlock()

				ch.hookLock.Lock()
				for _, hook := range ch.hooks {
					hook(res)
				}
				ch.hookLock.Unlock()
			}
		}
	}(checker)
	return nil
}

func (checker *AsyncHealthChecker) Stop() error {
	if checker.stopChan == nil {
		return ErrHealthcheckerNotStarted
	}
	checker.stopChan <- struct{}{}
	return nil
}

func (checker *AsyncHealthChecker) RegisterIndicator(indicator HealthIndicator) {
	checker.indicatorLock.Lock()
	checker.indicators[indicator.Name()] = indicator
	checker.indicatorLock.Unlock()
}

func (checker *AsyncHealthChecker) UnregisterIndicator(indicator HealthIndicator) {
	checker.indicatorLock.Lock()
	delete(checker.indicators, indicator.Name())
	checker.indicatorLock.Unlock()
}

func (checker *AsyncHealthChecker) AddHook(name string, hook HealthCheckerHook) {
	checker.hookLock.Lock()
	checker.hooks[name] = hook
	checker.hookLock.Unlock()
}

func (checker *AsyncHealthChecker) RemoveHook(name string) {
	checker.hookLock.Lock()
	delete(checker.hooks, name)
	checker.hookLock.Unlock()

}
