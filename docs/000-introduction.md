---
title: Introduction
---

# Introduction

[Meeseeks-box](https://github.com/pcarranza/meeseeks-box) is a ChatOps
Construction Kit that allows anyone to build your own automations following
the UNIX principle of using small tools that know how to do 1 thing right.

In the case of the Meeseeks-Box itself, it knows how to talk to Slack, listen
for messages and dispatch jobs to be executed as if it was being executed by
a user in bash.

The idea started when using [COG](https://github.com/operable/cog) to
automate tasks and finding it complex or hard to build automations that could
be simply scripted in bash.

## Driving ideas

### Simplicity

Simplicity is key to allow anyone to start automating tasks right away. Most
systems engineers are quite comfortable creating bash scripts as tools, these
bash scripts can be used from the chat, lowering the bar to automation and
enabling other people to operate systems remotely.

The drive for simplicity made me write this in Go, so installation is simply
downloading a binary (darwin, linux amd64 and armv6 are provided), exporting
a `SLACK_TOKEN` environment variable and start running it.

The drive for simplicity made command registration be setting a configuration
file that points at the executable, the permissions model and the arguments
that will be included. This allows using any available command in the running
box (echo, curl, or even docker) turning the Meeseeks-Box into a glue kind of
command, like bash.

The drive for simplicity made me use
[BoltDB](https://github.com/coreos/bbolt) as an embedded database to not have
any dependency whatsoever to persist data right away from the start.

### Security and safety

Some commands are more dangerous than others, so every time a command is
registered it will start with a AllowNone strategy, forcing the administrator
to pick what level of security to use, other options are AllowGroup and
AllowAny.

Some builtin commands (like the audit branch) uses AllowAdmin which requires
this group to be defined with the right users to enable using them.

Regarding execution of commands, these are being launched with the go
[os/exec](https://golang.org/pkg/os/exec/#CommandContext) package, setting a
60 seconds timeout by default to prevent commands from blocking forever,
eventually starving the box.

This package internally uses `os.StartProcess`, which evetually derives in
the system calls `fork` and `exec`. This is safer than the classic `system`
approach as it prevents a final user from injecting commands with colons into
the argument list.

Executed jobs are always recorded, the command, arguments, start and finish
dates are recorded, along with the outputed logs, both on stdout and stderr,
and the final error returned by the execution. These logs are available
throught the `jobs` and `logs` commands to list the executed jobs and show
the output, or `last` and `tail` if you just want to see the last recorded
job.

### Flexibility

Because any command can be used to build automations users are allowed to
build their ChatOps experience with the tools that are better suited for them.

So, you could:

- Invoke a remote API using `curl`
- Use a client tool such like `consul` or `kubectl` that will allow you to manage much more complex services.
- Create one docker container per command by using the `docker` command.
- Invoke bash scripts that perform complex operations.
- In the future, create specific clients using the GRPC API.

This is thought as glue, so you can build your own experience, according to your needs.

### Future proof

There are some future plans that will be soon implemented that will allow
this tool to be ready for any change or need that may come.

#### Execution Locallity

Sometimes you need your script, or tool, to be executed in a specific
location. Imagine that you have a fleet composed by many hosts, and that you
need to run a command on every of the hosts that are in a specific tier.
Now imagine that you have one Meeseeks-Box talking to Slack, and that you
have one Meeseeks-Box running on each of the hosts of the fleet, each one
with labels that identify where are they, or what their role is.

Now imagine that you can invoke a command to be executed in every host that
matches a label search from slack by issuing `@meeseeks command -labels
tier=frontend,type=web args` and that this will issue the command execution
local to the host that you need to be running the command.

No more complex and insecure ssh setups.

#### API invokations