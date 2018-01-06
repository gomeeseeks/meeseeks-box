# Meeseeks Box

> You make a request
>
> The Meeseek fulfils the request
>
> And then it stops existing

Meeseeks box is an automation engine built for simplicity, extensibility and
composability.

It follows the Unix philosophy of doing only one thing well, only caring about
gluing commands allowing everyone to build their own ChatOps experience by
going to the basics of using simple bash scripts as the execution unit.

# How to use

It is composed by a single go binary that requires a `SLACK_TOKEN` environment
variable set to start up.

To start adding commands add a configuration file and load it adding the
`-config=config-file.yml` argument and then restarting the binary (still no hot
reload supported)

## Example file

```yaml
commands:
  echo:
    command: "echo"
    auth_strategy: any
    timeout: 5
    help: command that prints back the arguments passed
```

## Command configuration

A command can be configured the following way:

- `command`: the command to execute
- `args`: list of arguments to always prepend to the command
- `timeout`: how long we allow the command to run until we cancel it, in
  seconds, 60 by default
- `auth_strategy`: defined the authorization strategy
  - `any`: everyone will be allowed to run this command
  - `none`: no user will be allowed to run this command (default value,
    permissions have to be explicit and conscious)
  - `group`: use `allowed_groups` to control who has access to this command
- `allowed_groups`: list of groups allowed to run this command
- `help`: help to be printed when using the builtin `help` command
- `templates`: adds the capacity to change how the replies from this command
  are represented, check the Templating section.

## Builtin Commands

Meeseeks include a set of builtin commands that can be used to introspect the
system, these are:

- help: prints the list of configured commands
- groups: prints the configured groups and which users are included there
- version: prints the current executable version

# Configuration

Besides configuring commands, other things can be configured

## Permissions

- `groups`: map group name and user list. By using the `allowed_groups` we will
  be forced to define which users are in which groups, this can only be set in
  configuration for now, for which

For example:

```yaml
groups:
  admin: # default admin group used by builtin commands
  - "user1"
  - "user2"
```

## Interface

It's strange that you want to change this, but anyway, here it is.

- `messages`: map that contains a list of strings, is used to build the
  meeseeks experience, default values can be checked in the
  `meeseeks/template/template.go` file.
- `colors`: colors to use for info, error and success messages in the slack
  interface.

# Templating

Commands support to change templates. All the templating is done with the go
package `text/template`, and all the rendering data is submitted with a Payload
that is nothing but a `map[string]interface{}`, handle with care, but if you
insist, check the following files to understand how it works:

- [Templates](./meeseeks/template/template.go)
- [Sample tests](./meeseeks/template/template.go)
