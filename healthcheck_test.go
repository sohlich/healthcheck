package healthcheck

import (
	"log"
	"testing"
	"time"
)

type TestIndicator struct{}

func (t *TestIndicator) Name() string {
	return "test"
}

func (t *TestIndicator) IsHealthy() bool {
	return true
}
func TestBasic(t *testing.T) {
	checker := New(time.Second)
	checker.RegisterIndicator(&TestIndicator{})
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
