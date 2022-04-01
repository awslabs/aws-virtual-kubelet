# Objects (types) in this package

**`TODO`** Add description/diagram of flow

## Summary
- `PodMonitor` configures and runs pod-related monitors
- `Monitor` holds details about a specific monitor and provides a way to run checks
- `Watcher` is an _add-on_ to a standard check-based (polling) monitor that replaces check behavior
- `CheckHandler` is a type of `Handler` that receives `CheckResults` and determines what to do with them
- `checkResult` is the result of running a check or receiving a response from a watch-based monitor

# Functions
`checks.go` contains the implementations for check/watch functions referenced by a `Monitor` and invoked by `PodMonitor` (these functions return a `checkResult`)

## Details

### `PodMonitor`
The `PodMonitor` creates and configures a set of monitors appropriate for monitoring a `Pod`.  When started, it will being running each monitor's checks or start watch-based monitor's receive loop.  Results of the monitoring checks/watches are sent to the configured `CheckHandler` (which is passed as an argument when creating the `PodMonitor`).  These results are passed via Go channels to allow the `Handler` to process them sequentially and avoid loops / issues arising from processing multiple results at the same time.

`PodMonitor`s also expose a (boolean) channel so the main monitoring loop can be signalled to stop.  This signal is passed along to any watch-based stream receiver loops, so they also stop processing.

### `Monitor`
A `Monitor` holds information about what is being monitored (`Subject`), the thing affected by monitoring states (`Resource`), and how to monitor it (initiating polling checks or receiving streaming watch results).  `MonitorState` and `Subject` types represent the monitor's state and subject (and _belong to_ a `Monitor`).  `Resource` is a generic type associated with a monitor for use when handling monitoring states. `Monitor`s track check failures and determine their state based on failure counts (vs. the configured failure threshold).  For watch-based monitors the remote system is expected to provide appropriate status information, which is mapped to a monitoring state by the watch implementation.

A `Monitor`s `Dependency` member points to a _parent_ monitor.  This relationship indicates that the _parent_ must be _Healthy_ before the current monitor can be healthy[^1].  This allows remediation behavior to take efficient action when multiple monitors are _Unhealthy_.  For instance if an EC2 instance is unhealthy, it is unlikely that a service running on that instance will be healthy.

### `Watcher`
The `Watcher` _add-on_ to a monitor replaces check (polling) behavior and further configures the monitor to receive streamed health check results from a remote system.  In addition to storing a reference to the watch function's implementation, a `Watcher` provides a (boolean) channel to allow signaling the asynchronous receiver loop to stop processing.

### `checkResult`
`checkResult`s are the result of a check/watch invocation and are returned for processing by check/watch functions.  They hold details about the particular check result as well as a reference to the `Monitor` they belong to, a timestamp capturing when the result was obtained, and a generic `Data` member to hold any information returned with the result[^2].  They are sent directly to the handler by the monitoring orchestrator (`PodMonitor` in this case).

### `CheckHandler`
A `CheckHandler` receives check/watch results and determines appropriate action.  Generally it will try to restart or remediate failing monitor subjects and, if unsuccessful, will destroy and recreate necessary resources to effect a _restart_.  

If attempts to destroy and recreate the affected resource fail, then the handler will try the same remediation on any _parent_ (`Dependency`) declared by the monitor.  If no dependency is set (indicating we are at a root dependency), then the handler will continue to retry the recreation attempts indefinitely.  This remediation cycle can loop forever in some failure scenarios and should be detected by an outside system that is watching related metrics[^3].

`CheckHandler`s also process non-failing check results (which immediately result in a _Health_ monitor) and know how to process the `Data` for particular result and potentially update the `Resource` of the associated monitor.

### Check Functions
Check / watch functions should not return errors.  They should only return a `CheckResult` (if an unexpected error occurs, it's still a failed check).

[^1]: See the [monitor.puml](monitor.puml) [PlantUML](https://plantuml.biz) diagram for an example.
[^2]: This capability is currently used to get the `PodStatus` from a watch-based Application Lifecycle check, but can hold any type.
[^3]: **`TODO`** document likely metrics to monitor
---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.