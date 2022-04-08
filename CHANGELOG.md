# Changelog

## [v0.5.2+opensource]() (2022-04-06)
### 🎉 Features
* **ec2:** changed EC2 API polling loop to use [Waiters](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/making-requests.html#using-waiters)

### 🧹 Chores
* **docs:** added verbiage, templates, and files for Open Source version
* **cleanup:** removed **`TODO`**s, comments, and unused code throughout

## [v0.5.2]() (2022-03-23)

### 🐛 Bug Fixes
* **awsutils:** log an Error if EC2 `RunInstances` doesn't return an instance ID
* **awsutils:** fix missing `instance-id` annotation on EC2 instances in some VK restart cases
* **health:** include non-running instances in EC2 health check (previously would only validate Running instances)
* **health:** logic refactor to avoid stacking goroutines and remove unnecessary sleeps
* **provider:** CreatePod logic refactor (mirrors DeletePod now)...should resolve premature Application Health Watch

### 🎉 Features
* **docs:** added [CHANGELOG.md](CHANGELOG.md) and template w/ examples
* **docs:** added [EdgeCases.md](docs/EdgeCases.md) to capture potential (unhandled) edge-cases and ideas to mitigate
* **docs:** added [BacklogFodder.md](docs/BacklogFodder.md) to capture possible future backlog items
* **provider:** moved Pod notifier to utils so other packages could access without import loops
* **warm-pool:** feature is available again (add a `WarmPool` config to enable or remove to disable)

### 🧹 Chores
* **ec2_utils:** update to klog v2
* **logging:** remove large struct dumps and improve verbose vs. not log distinction

### ❗️ Known Issues
* **tests:** unit tests are failing due to unexpected output from the vkvmagent

---

# Template
Use this template and checklist to create new changelog entries.

- [ ] update `v#.#.#` with the new revision
- [ ] insert the diff URL comparing against the previous version
- [ ] update the `(date)`
- [ ] note new _Features_, _Bug Fixes_, etc. using the formatting examples provided below (**`TIP`** compare the previous version tag with `HEAD` to see new PRs / changes)

##  **`WIP`** 🚧 [v0.0.2]() (2021-01-02)
> Next Release _Work In Progress_

## [v0.0.1]() (2021-01-01)
> Deployed Release

**`NOTE`** optionally add [#00]() pull request reference/link after scope.

### 🎉 Features
* **scope:** brief description of the new feature

### 🐛 Bug Fixes
* **scope:** brief description of the problem and resolution

### ↩️  Reverts
* **scope:** revision(s) reverted and reason

### 🧹 Chores
* **scope:** explanation of chore/task completed and reason

### ❗️ Known Issues
* **scope:** explanation of issue and plan to resolve