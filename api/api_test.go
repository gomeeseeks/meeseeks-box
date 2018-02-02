package api_test

import (
	"testing"
)

/*
# API

## Will need

* a token that links to a user+command+channel
* a command to create a token
* a command to revoke a token
* a command to list tokens with commands

## Steps to enable the API

1. Configuration: the endpoint in which to listen, namely address and path.
2. Enabling: using a command, pass in the command that the API will be
calling and the channel in which to reply, else it will be considered as an
IM
3. The command will need to resolve the channel <token> (the client will need to provide this)
4. This command can only be called over IM
5. ~The api enabling will need to test that the user has access to the requested command?~ (nah... late binding)

## Token data model

- Tokens
  - Token (something something UUID hash)
    - UserID
    - ChannelID
    - Command
    - Args

### Interface

- Create{payload} - returns the token
- Get{Token}
- List{UserID}

# Steps to build:

1. Create the data model
2. Add the creation and the Get methods
3. Add the http interface to post the token, return 200 when found
4. Pipe the message to the stub client, check it works
5. Finish implementation

*/

func TestSomething(t *testing.T) {
}
