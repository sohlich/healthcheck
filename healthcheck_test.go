package healthcheck

import (
	"log"
	"sync"
	"testing"
	"time"
)

type TestIndicator struct {
	*sync.Mutex
	hasBeenCalled bool
}

func (t *TestIndicator) Name() string {
	return "test"
}

func (t *TestIndicator) IsHealthy() bool {
	t.Lock()
	t.hasBeenCalled = true
	t.Unlock()
	return true
}

func TestBasicFunction(t *testing.T) {
	checker := New(time.Millisecond)
	checker.RegisterIndicator(&TestIndicator{
		&sync.Mutex{},
		true,
	})
	cond := make(chan bool, 0)

	checker.AddHook("testHook", func(m map[string]bool) {
		ok := m["test"]
		select {
		case cond <- ok:
			log.Println("Condition has been written")
		default:
			log.Println("Cant write condition")
		}
	})
	checker.Start()
	if !<-cond {
		t.Error("Helthchecker failed")
	}
	checker.Stop()
}

func TestStopIfNotStarted(t *testing.T) {
	checker := New(time.Millisecond)
	err := checker.Stop()
	if err == nil {
		t.Error("Stopped not started checker")
	}
}

func TestStartAlreadyStartedChecker(t *testing.T) {
	checker := New(time.Millisecond)
	err := checker.Start()
	if err != nil {
		t.Error("Start checker failed")
	}
	err = checker.Start()
	if err == nil {
		t.Error("The already started checker cannot be started")
	}
	checker.Stop()
}

func TestUnregister(t *testing.T) {
	checker := New(time.Millisecond)
	indicator := &TestIndicator{
		&sync.Mutex{},
		false,
	}
	checker.RegisterIndicator(indicator)
	checker.Start()
	time.Sleep(2 * time.Millisecond)
	checker.UnregisterIndicator(indicator)
	indicator.Lock()
	indicator.hasBeenCalled = false
	indicator.Unlock()
	time.Sleep(2 * time.Millisecond)

	if indicator.hasBeenCalled {
		t.Error("Unregistration not sucessfull")
	}
}

func TestRemoveHook(t *testing.T) {
	checker := New(time.Millisecond)

	counter := struct {
		*sync.Mutex
		called bool
	}{
		&sync.Mutex{},
		false,
	}

	checker.AddHook("testHook", func(m map[string]bool) {
		counter.Lock()
		counter.called = true
		counter.Unlock()
	})
	checker.Start()
	time.Sleep(2 * time.Millisecond)
	counter.Lock()
	if !counter.called {
		t.Error("Counter not called")
	}
	counter.called = false
	counter.Unlock()
	checker.RemoveHook("testHook")
	time.Sleep(2 * time.Millisecond)
	counter.Lock()
	if counter.called {
		t.Error("Counter should not be called")
	}
	counter.Unlock()
	checker.Stop()
}
