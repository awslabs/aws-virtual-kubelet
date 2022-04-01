# Contributing Guidelines

Thank you for your interest in contributing to our project. Whether it's a bug report, new feature, correction, or
additional documentation, we greatly value feedback and contributions from our community.

Please read through this document before submitting any issues or pull requests to ensure we have all the necessary
information to effectively respond to your bug report or contribution.

To get an overview of the project, see the [README](../README.md) doc.

## Reporting Bugs/Feature Requests

We welcome you to use the GitHub issue tracker to report bugs or suggest features.

When filing an issue, please check existing open, or recently closed, issues to make sure somebody else hasn't already
reported the issue. Please try to include as much information as you can. Details like these are incredibly useful:

* A reproducible test case or series of steps
* The version of our code being used
* Any modifications you've made relevant to the bug
* Anything unusual about your environment or deployment

## Contributing via Pull Requests

Contributions via pull requests (PRs) are much appreciated. Before sending us a PR, please ensure that:

1. You are working against the latest source on the *main* branch.
2. You check existing _open_ and _recently merged_ PRs to make sure someone else hasn't addressed the problem already.
3. You open an issue first to discuss any significant work before spending time making changes.

To send us a PR, please:

1. Fork the repository.
2. Modify the source; please focus on the specific change you are contributing. If you also reformat all the code, it
   will be hard for us to focus on your change (**`TIP`** Create a separate PR with just the formatting changes for easier review).
3. Ensure local tests pass.
4. Commit to your fork using clear commit messages (https://www.conventionalcommits.org/ style is strongly encouraged)
5. Send us a PR, answering any default questions in the PR interface.
6. Pay attention to any automated CI failures reported in the PR, and stay involved in the conversation.

GitHub provides additional document on [forking a repository](https://help.github.com/articles/fork-a-repo/) and
[creating a pull request](https://help.github.com/articles/creating-a-pull-request/).

### Recommended workflow
**`TODO`** **Currently this isn't a good approach since PRs from forks aren't running any workflows** 😕

It is recommended to add this repository as an _upstream_ remote in your checkout, and to create branches off this
repository's _main_ branch. e.g.

_In a checkout of your fork add this repository as an upstream remote_

```shell
git remote add upstream git@github.com:aws/aws-virtual-kubelet.git
```

**`NOTE`** other URL types such as HTTPS should work also.

_Create (and switch to) a local branch tracking this repository's _main_ branch_ (upstream repo primary branch)

```shell
 git checkout -b upstream-master upstream/master
```

_Now create (and switch to) a new PR branch off the upstream primary for your PR_

```shell
git checkout -b feat/my-awesome-pull-request
```

Before starting a new PR branch, `git pull` while switched to the upstream primary branch created above to ensure you
are working with the latest code.

This approach simplifies the review process by ensuring that only relevant commits show up in the PR. It also allows
your local fork's own _main_ branch to have differences without accidentally merging those into the main repo (
e.g. status badge URLs).

### Additional tips

Use [Conventional Commits](https://www.conventionalcommits.org/) style commit comments to help organize your change
sets. The [git book](https://git-scm.com/book) recommends keeping your work local until you're ready to share.

**`NOTE`** When PRs are merged into the primary repo all commits
are "[squashed](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github#squashing-your-merge-commits)"
into a single commit to encourage a concise commit history.

Make judicious use of `git commit --amend` and other [history rewriting](https://git-scm.com/book) capabilities of git _
before sharing your code_ to provide a clear and succinct set of changes for PR reviewers. See
the [Cookbook](docs/Cookbook.md#version-control) for additional details.

## Finding contributions to work on

Looking at the existing issues is a great way to find something to contribute on. As our projects, by default, use the
default GitHub issue labels (enhancement/bug/duplicate/help wanted/invalid/question/wontfix), looking at any 'help
wanted' issues is a great place to start.

## Code of Conduct

This project has adopted the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct). For more
information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq) or contact
opensource-codeofconduct@amazon.com with any additional questions or comments.

## Security issue notifications

If you discover a potential security issue in this project we ask that you notify AWS/Amazon Security via
our [vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/). Please do **not** create a
public github issue.

## Licensing

See the [LICENSE](LICENSE) file for our project's licensing. We will ask you to confirm the licensing of your
contribution.

---
>© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
This work is licensed under a Creative Commons Attribution 4.0 International License.