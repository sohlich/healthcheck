//Package healthcheck provides simple implementation of application
//periodic health checker. That periodically checks the health of application
//compoentns.
package healthcheck

import (
	"errors"
	"sync"
	"time"
)

//Errors that are thrown if HealthChecker is in illegal state.
var (
	//ErrHealthcheckerNotStarted is returned by HealthChecker.Stop() method if
	//the checker has not been started yet.
	ErrHealthcheckerNotStarted = errors.New("HealthChecker not started yet.")
	//ErrHealthcheckerAlreadyStarted is returned by HealthChecker.Start() method
	//if the checker has been already started.
	ErrHealthcheckerAlreadyStarted = errors.New("HealthChecker already started.")
)

//HealthIndicator is the method the wraps the function that check the
//given application component.
//
//Health indicator should not been long running proccess, because the
// indicators method IsHealthy is called in infinite loop of HelathChecker
//and that could block othe health indicators to be callled for a long time.
type HealthIndicator interface {

	//Name should return
	//the name of indicator.
	Name() string

	//IsHealthy return true if compoent
	//is healthy and false in oposit case.
	IsHealthy() bool
}

//HealthCheckerHook is a function that is called after
//all health checks are done.
type HealthCheckerHook func(map[string]bool)

//HealthChecker interface wraps the functionality
//of package. After the HealthChecker is started the infinite loop is
//started and ii iterates over all regitered HealthIndicators. After the health check
//is done, the result is then submitted to added hooks.
//The HealthChecker  is considered to be thread safe.
type HealthChecker interface {
	//Start is called to start the infinite loop
	//to check the HealthIndicators.
	Start() error

	//Stop stops the HealthChjeckers
	//infinite loop.
	Stop() error

	//RegisterIndicator register the indicator
	//to HelathChecker object.
	RegisterIndicator(indicator HealthIndicator)

	//UnregisterIndicator removes the HealthIndicator
	//from the collection of checked inidcators.
	UnregisterIndicator(indicator HealthIndicator)

	//AddHook adds a hook to the collection of hooks to be called
	//after all checks are done.
	AddHook(name string, hook HealthCheckerHook)

	//RemoveHook removes the hook by given name.
	RemoveHook(name string)
}

type asyncHealthChecker struct {
	period        time.Duration
	indicatorLock *sync.Mutex
	hookLock      *sync.Mutex
	indicators    map[string]HealthIndicator
	hooks         map[string]HealthCheckerHook
	stopChan      chan struct{}
}

//New creates new HealthChecker object.
func New(period time.Duration) HealthChecker {
	checker := &asyncHealthChecker{
		period:        period,
		indicatorLock: &sync.Mutex{},
		hookLock:      &sync.Mutex{},
		indicators:    make(map[string]HealthIndicator),
		hooks:         make(map[string]HealthCheckerHook),
	}

	return checker
}

func (checker *asyncHealthChecker) Start() error {
	if checker.stopChan != nil {
		return ErrHealthcheckerAlreadyStarted
	}

	checker.stopChan = make(chan struct{}, 0)
	go func(ch *asyncHealthChecker) {
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
					go func(r map[string]bool) {
						hook(r)
					}(res)
				}
				ch.hookLock.Unlock()
			}
		}
	}(checker)
	return nil
}

func (checker *asyncHealthChecker) Stop() error {
	if checker.stopChan == nil {
		return ErrHealthcheckerNotStarted
	}
	checker.stopChan <- struct{}{}
	return nil
}

func (checker *asyncHealthChecker) RegisterIndicator(indicator HealthIndicator) {
	checker.indicatorLock.Lock()
	checker.indicators[indicator.Name()] = indicator
	checker.indicatorLock.Unlock()
}

func (checker *asyncHealthChecker) UnregisterIndicator(indicator HealthIndicator) {
	checker.indicatorLock.Lock()
	delete(checker.indicators, indicator.Name())
	checker.indicatorLock.Unlock()
}

func (checker *asyncHealthChecker) AddHook(name string, hook HealthCheckerHook) {
	checker.hookLock.Lock()
	checker.hooks[name] = hook
	checker.hookLock.Unlock()
}

func (checker *asyncHealthChecker) RemoveHook(name string) {
	checker.hookLock.Lock()
	delete(checker.hooks, name)
	checker.hookLock.Unlock()

}
