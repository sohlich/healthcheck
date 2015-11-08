package healthcheck_test

import (
	"log"
	"net/http"
	"time"

	"github.com/sohlich/healthcheck"
)

// SimpleHttpAliveIndicator test the availability of given
// url address.
type SimpleHttpAliveIndicator struct {
}

func (t *SimpleHttpAliveIndicator) Name() string {
	return "httpAlive"
}

func (t *SimpleHttpAliveIndicator) IsHealthy() bool {
	err, _ := http.Get("http://google.com")
	if err != nil {
		return false
	}
	return true
}

// SimpleLogHook logs the result of health checks via builtin
// log package
func SimpleLogHook(m map[string]bool) {
	if m["httpAlive"] {
		log.Println("The service http://google.com is reachable")
	} else {
		log.Println("The service http://google.com is unreachable")
	}
}

//Basic usage example.
func Example() {
	// Create healthchecker
	checker := healthcheck.New(time.Millisecond)

	// Register HealthIndicator
	checker.RegisterIndicator(&SimpleHttpAliveIndicator{})

	// Add hooks that consume the results.
	checker.AddHook("testHook", SimpleLogHook)

	// Start the checker
	checker.Start()
	time.Sleep(10 * time.Second)

	// Before the application finish try
	// to gracefully shutdown the checker
	checker.Stop()
}
