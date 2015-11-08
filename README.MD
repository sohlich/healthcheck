# Go healthchecker

This package provide basic tool for application components health checking.


## Usage

The healthchecker object use the HealthIndicators to obtain the health status of
a component. The indicators are executed sequentialy in infinite loop to keep
the complete status updated. The indicators should not be long running processes/functions.

```
   // Create healthchecker
    checker := healthcheck.New(time.Millisecond)

    // Register HealthIndicator
    checker.RegisterIndicator(&SimpleHttpAliveIndicator{})

    // Add hooks that consume the results.
    checker.AddHook("testHook", SimpleLogHook)

    // Start the checker
    checker.Start()
```