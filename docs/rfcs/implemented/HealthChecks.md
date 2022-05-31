This document describes the behavior of the health checking / monitoring system.

## Summary
At a high level, there is one `PodMonitor` associated with each pod.  That pod monitor manages one or more monitors which observe and track the state of pod components/resources.  When those observations (`checkResults`) record failures, a per-monitor failure count is incremented.  When that count reaches a configurable threshold, the monitor is said to be `Unhealthy`.  A `CheckHandler` processes every check result (healthy or not) and contains logic to respond to the unhealthy state.  Any successful check results in reset of the failure counter and a return to the `Healthy` state for the monitor.

It is important to note that there are 2 kinds of monitor observations: Check (polling) and Watch (streaming receive).  These are mirror images of each other but have the same end result of providing observations about resources that increment or reset failures which affect a monitor's state.  Observations mostly run in parallel via goroutines, but check result handling is sequential via channels.  This helps ensure orderly processing of state and minimizes aberrant behavior.

## Objects (types) in this package

- `PodMonitor` configures and runs pod-related monitors
- `Monitor` holds details about a specific monitor and provides a way to run checks
- `Watcher` is an _add-on_ to a standard check-based (polling) monitor that replaces check behavior
- `CheckHandler` is a type of `Handler` that receives `CheckResults` and determines what to do with them
- `checkResult` is the result of running a check or receiving a response from a watch-based monitor

## Functions
`checks.go` contains the implementations for check/watch functions referenced by a `Monitor` and invoked by `PodMonitor` (these functions return a `checkResult`)

## Details

### `PodMonitor`
The `PodMonitor` creates and configures a set of monitors appropriate for monitoring a `Pod`.  When started, it will begin running each monitor's checks or start watch-based monitor's receive loop.  Results of the monitoring checks/watches are sent to the configured `CheckHandler`.  These results are passed via Go channels to allow the `Handler` to process them sequentially and avoid loops / issues arising from processing multiple results at the same time.

`PodMonitor`s create a cancellable context when started.  This context is passed to all monitors to allow single-point cancellation of any goroutines started by monitor checks/watches.  WaitGroups are also used to track goroutines to help ensure leakage does not occur.

### `Monitor`
A `Monitor` holds information about what is being monitored (`Subject`), the thing affected by monitoring states (`Resource`), and how to monitor it (initiating polling checks or receiving streaming watch results).  `MonitorState` and `Subject` types represent the monitor's state and subject (and _belong to_ a `Monitor`).  `Resource` is a generic type associated with a monitor for use when handling monitoring states. `Monitor`s track failures and determine their state based on failure counts (vs. the configured failure threshold).  For watch-based monitors the remote system is expected to provide appropriate status information, which is mapped to a monitoring state by the watch implementation.

### `checkResult`
`checkResult`s are the result of a check/watch invocation and are returned for processing by check/watch functions.  They hold details about the particular check result as well as a reference to the `Monitor` they belong to, a timestamp capturing when the result was obtained, and a generic `Data` member to hold any information returned with the result[^1].  They are sent directly to the handler by the monitoring orchestrator (`PodMonitor` in this case).

### `CheckHandler`
A `CheckHandler` receives check/watch results and determines appropriate action.

`CheckHandler`s also process non-failing check results (which immediately result in a _Healthy_ monitor) and know how to process the `Data` for particular result and potentially update the `Resource` of the associated monitor.

### Check Functions
Check / watch functions should not return errors.  They should only return a `CheckResult` (if an unexpected error occurs, it's still a failed check).

[^1]: This capability is currently used to get the `PodStatus` from a watch-based Application Lifecycle check, but can hold any type.