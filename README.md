# Meeseeks Box

> You make a request
>
> The Meeseek fulfils the request
>
> And then it stops existing

# Working backwards

After the massive success of the butterbot we need to get more ambitious and cover more use cases. For this I will explain roughly how does a flexible chatops system looks like from my operational perspective, and how does the design looks like from my development perspective.

## Things to have

- Flexibility in running whatever we want to run, probably going down to simply running scripts for maximum flexibility.
- Security both at the front end with a permissions model (that can be based on the slack username) and at the backed using 2FA for configured jobs.
- Reliability from the perspective of the system supporting any part going dark and not impacting the other.
- Scalability in that we can horizontally scale it with the fleet.
- Persistence, we need to be able of getting data from the past.
- Humor, in that it's fun to use the system, and to extend it.
- Simplicity, in that the system is really simple to build, and operate.
- Visibility, in that the system is fully monitored and provides metrics with prometheus.
- Routability, in that there are times in which data locality matters, a lot.
- Specialization, in that we can start building highly specialized parts of the system, for kubernetes, for example.

## Constraints

- Only make this available with Slack, at least initially.
- Detach execution (and state) from interface. Slack has a hard limit of 3 seconds to get a reply, we can't assume all commands are going to be this fast.
- Separate orchestration and execution. We can have many execution points, we should only have one orchestration point (even if HA)

## Ok, but how do we do this?

I propose that we start with the butterbot and start adding parts to the system with different focuses along the way.

### Design Stages

#### Enter Meeseeks Box

Picking the Butterbot, we add a configuration file that allows us to setup multiple jobs, thus the execution looks more like this model:

> user: @mrmeeseeks
> mrmeeseeks: I'm Mr Meeseeks! look at me! \n <list of tasks it can do loaded from [commands list] in the configuration file>
> user _picks the action to perform_
> mrmeeseeks _updates the message with further prompting on details or arguments_
> user _picks the argument_
> mrmeeseeks _promtps the user if he's sure_
> user _clicks on the yes button_
> mrmeeseeks: Uuuuh, yeah! can do!
> ... (execution happens)
> mrmeeseeks: @user All done! \n ```results of the command, if any```

#### Enter Remote Meeseeks

We add one more layer to the system and _detach execution from orchestration_. We simply create an independent MrMeeseeks binary that gets the URL of the MeeseeksBox and using GRPC it starts long polling waiting for actions.

To connect the MeeseeksBox with one MrMeeseeks we define a shared token that gets stored in the configuration file.

At this stage we only allow one MrMeeseeks execution process, for simplicity.

The MeeseeksBox works as a passthrough

#### Increase complexity slowly

##### Persistence and state tracking

We add some form of database on the Meeseeks box and we make the messages persistent. We use this to also track registered Meeseeks.

We add labels to the Meeseeks, which we can later use for multiple things.

##### Commands catalog

Pretty much hand in hand with the previous task, we start storing which are the different capacities of the different Meeseeks, we use a simple random or round robin to send tasks to different Meeseeks

##### Dynamic Security Model

We start building a security model that would allow us to segregate users in groups or roles (TBD), then we can allow different roles to be allowed to run commands in selectors, for which we will use the labels.

##### Queueing of messages

We start allowing the queueing of requests by storing them and executing them whenever there's available capacity.

##### Enter Meeseek Multiplexing

We allow registering multiple Meeseeks. We add labels to these Meeseeks and we allow routing to specific ones by using a similar pattern to kubernetes: `-l environment=prd,type=X,app=Y`, this way we can route requests to specific places by name, or we can broadcast requests to many Meeseeks

Wiring this to the security model we can segregate permissions such that a given group can run a command only on some labels, which can be used to control environment access.

##### 2FA

For some jobs we can introduce 2FA which would allow a Meeseek to challenge a command with a 2FA request, this way we can make some specific commands high security because we are not going to only be relying on Slack to authenticate the user.
