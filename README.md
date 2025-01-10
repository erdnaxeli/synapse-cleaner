# Synapse cleaner

This project is intented at people using Synapse for their personal use only, that
is when they are the only local user of the server.

## Installation

Clone this repository, go the the folder, and then:
```shell
make build
```

A new binary `synapse-cleaner` should be present.

## Commands

### Purge rooms

This commands purge all rooms in which the given user is not present.

```
./synapse-cleaner purge-rooms --user @myuser:example.org --server myserver.example.org
```

As this command is destructive, it will show you how many rooms this user is present in,
the number of rooms it is about to purge, and will ask for confirmation.
