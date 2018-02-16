# Meeseeks Box

> You make a request
>
> The Meeseek fulfills the request
>
> And then it stops existing

[Meeseeks-box](https://github.com/gomeeseeks/meeseeks-box) is a ChatOps Construction Kit that allows anyone to build your own automations following the UNIX principle of using small tools that know how to do 1 thing right.

In the case of the Meeseeks-Box itself, it knows how to talk to Slack, listen for messages and dispatch jobs to be executed as if it was being executed by a user in bash.

The core tenets of the tool are simplicity, Security, and Flexibility.


## FAQ

> So... what is this?

The meeseeks are a way of running any executable in a host, through Slack, while keeping things simple and secure.

They are born from the fact that the world isn't pretty, and as much as we would like to have great elegant resilient systems, sometimes we just need to run some bash scripts (or curl, or whatever) somewhere in the fleet to keep things working.

So, instead of building ambitious projects that will never be coded, let alone deployed, you can start automating your toil away right now.

> What do I need to start using the meeseeks and automate my toil away?

Downloading a single binary file, and a slack api token

> What language can I use to automate my toil?

Any language, the meeseeks run commands using `fork+exec`, you can use anything that can be executed by a shell.

> What is the command API that I have to implement in my scripts?

None, or better said, just POSIX, write whatever you want to stdout, that's the text that will be transported back to the chat. Return an error code different than 0 and it will be a command failing.

> Can I have long running commands? What sort of timeout commands have?

Yes, the meeseeks are built for an imperfect world in which things can take a long time. Default timeout is 60 seconds, but it can be configured per command, without a limit.

> Can I kill a command while it's running?

Yes, you can cancel the ones you own with `cancel job_id`, and admins can cancel any job with `kill job_id`. This will send a kill signal to the running command.

> Can I see the output of a command while it's running?

Yes, use `tail` to show the last 5 output lines from the last command that you launched.

## Documentation

For more in depth details, check the [docs](https://gomeeseeks.github.io/meeseeks-box/).
