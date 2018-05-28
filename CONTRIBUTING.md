# How to Contribute

Scope is [Apache 2.0 licensed](LICENSE) and accepts contributions via GitHub
pull requests. This document outlines some of the conventions on development
workflow, commit message formatting, contact points and other resources to make
it easier to get your contribution accepted.

We gratefully welcome improvements to documentation as well as to code.

# Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

# Email, Chat and Community Meetings

The project uses the the scope-community email list and Slack:
- Email: [scope-community](https://groups.google.com/forum/#!forum/scope-community)
- Chat: Join the [Weave community](https://weaveworks.github.io/community-slack/) Slack workspace and use the [#scope](https://weave-community.slack.com/messages/scope/) channel

When sending email, it's usually best to use the mailing list. The maintainers are usually quite busy and the mailing list will more easily find somebody who can reply quickly. You will also be potentially be helping others who had the same question.

We also meet regularly at the [Scope community meeting](https://docs.google.com/document/d/103_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/). Don't feel discouraged to attend the meeting due to not being a developer. Everybody is welcome!

## Getting Started

- Fork the repository on GitHub
- Read the [README](README.md) for getting started as a user and learn how/where to ask for help 
- If you want to contribute as a developer, continue reading this document for further instructions
- Play with the project, submit bugs, submit pull requests!

## Contribution workflow

This is a rough outline of how to prepare a contribution:

- Create a topic branch from where you want to base your work (usually branched from master).
- Make commits of logical units.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- If you changed code:
   - add automated tests to cover your changes
- Submit a pull request to the original repository.

## How to build and run the project

```bash
make
./scope launch
```

## How to run the test suite

You can run the linting and unit tests by simply doing

```bash
make tests
```

There are integration tests for Scope, but unfortunately it's hard to set them up in forked repositories and the setup is not documented. Help is needed to improve this situation: https://github.com/weaveworks/scope/issues/2192

# Acceptance policy

These things will make a PR more likely to be accepted:

 * a well-described requirement
 * tests for new code
 * tests for old code!
 * new code and tests follow the conventions in old code and tests
 * a good commit message (see below)

In general, we will merge a PR once two maintainers have endorsed it.
Trivial changes (e.g., corrections to spelling) may get waved through.
For substantial changes, more people may become involved, and you might get asked to resubmit the PR or divide the changes into more than one PR.

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various
git tools.

## 3rd party plugins
So you've built a Scope plugin. Where should it live?

Until it matures, it should live in your own repo. You are encouraged to annouce your plugin at the [mailing list](https://groups.google.com/forum/#!forum/scope-community) and to demo it at a [community meetings](https://docs.google.com/document/d/103_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/).

If you have a good reason why the Scope maintainers should take custody of your
plugin, please open an issue so that it can potentially be promoted to the [Scope plugins](https://github.com/weaveworks-plugins/) organization.
