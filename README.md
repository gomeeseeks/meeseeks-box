# Meeseeks Box

[![Build Status](https://travis-ci.org/gomeeseeks/meeseeks-box.svg?branch=master)](https://travis-ci.org/gomeeseeks/meeseeks-box)

[![Go Report Card](http://goreportcard.com/badge/github.com/gomeeseeks/meeseeks-box)](https://goreportcard.com/report/github.com/gomeeseeks/meeseeks-box)

> You make a request
>
> The Meeseek fulfills the request
>
> And then it stops existing

[Meeseeks](https://github.com/gomeeseeks/) is a ChatOps Construction Kit that allows anyone to build their own automations following the UNIX principle of using small tools that know how to do one thing right.

Meeseeks-Box is the component that knows how to talk to Slack, listen for messages and dispatch jobs to be executed as if it was being executed by a user in a shell.

The core tenets of the tool are simplicity, security and flexibility.


## FAQ

###So... what is this?

Meeseeks is a way of running any executable on a host through Slack while keeping things simple and secure.

The project is based on the fact that server infrastructures aren't pretty. As much as we would like to have great elegant, resilient systems, oftentimes all we really need is to simply run some shell scripts (or curl, or a db query, or whatever) somewhere in the fleet to perform a job.

So, instead of building ambitious projects that will never reach stability (let alone be deployed) you can start automating your toil away right now.

###What do I need to start using the Meeseeks and automate my toil away?

Download a single binary file and create a Slack API token. That's pretty much it.

###What languages can I use to automate my toil?

Any language.

The Meeseeks run commands using `fork+exec` so you can use anything that can be executed from a shell.

###What is the command API that I have to implement in my scripts?

None. Or better said, POSIX.

Write what you want to read to stdout: that's the text that will be transported back to the chat. Returning an exit code different than 0 will be interpreted as a command failure but the output will still be transported back.

###Can I have long running commands? What sort of timeout do commands have?

The Meeseeks are built for an imperfect world in which things can take a long time. The default timeout is 60 seconds but it can be configured on a per command basis. You can even spawn commands without a time limit.

###Can I kill a command while it's running?

Yes. You can cancel your own jobs with `cancel job_id`. Admins can cancel any job with `kill job_id`: this will send a kill signal to the running command.

###Can I see the output of a command while it's running?

Yes. Use `tail` to show the last output lines from the last command that you launched.

## Documentation

For more in depth details, check the [docs](https://gomeeseeks.github.io/meeseeks-box/).
