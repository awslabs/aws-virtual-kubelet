# Cookbook
This document contains a "cookbook style" reference of useful commands.  Each entry is a mini [runbook](https://wa.aws.amazon.com/wat.concept.runbook.en.html) of sorts designed to accomplish a specific task.  Collecting these in one place allows other documents to be more concise (linking here when appropriate) and provides a quick reference for later use.

Entries below are divided into sections by major topic area.

**`NOTE`** changing entry titles will likely break links from other documents so _find and replace all_ when doing so.

## AWS Deployment
### Build Virtual Kubelet and push to an ECR registry

```bash
export REGISTRY_ID=$MY_AWS_ACCOUNT_ID
export REGION=$MY_AWS_REGION
export GOOS=linux
make push
```

## Version Control
These entries pertain to `git` version control configuration and manipulation.  At the time of this writing the most recent version is `2.33` and has been tested with the steps below.  It's recommended to use the latest version when possible.

Installation instructions and download links can be found in the [git book download page](https://git-scm.com/downloads).

### Rewrite / organize commit history
⚠️ Use this with caution as it can cause loss of work (make a copy first if unsure)

#### amend the previous commit with additional changes and/or commit message updates
`git commit --amend`

#### manipulate the commit history in the current branch since it was created

This command creates an interactive rebase session that allows you to combine (squash) commits, modify commit messages, reorder commits, etc.

The `merge-base` command is a quick way to reference this branch's parent commit.  You can also specify a particular commit hash, or a relative reference such as `HEAD~3`.  See [Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History) in the [git book](https://git-scm.com/book) for more information.

git rebase -i `git merge-base HEAD <parent-branch-name>`

**`NOTE`** If you've already pushed work affected by the above commands to a remote repository, you will need to "force push" via `git push -f` to force rewriting history on the remote.  This should _**only**_ be done if you're not sharing this branch with anyone else, as it could cause others to lose changes unexpectedly.

See [Rebasing](https://git-scm.com/book/en/v2/Git-Branching-Rebasing) for additional information and [Learn Git Branching](https://learngitbranching.js.org/) for an excellent visual tutorial of both basic and advanced concepts and scenarios.

---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.