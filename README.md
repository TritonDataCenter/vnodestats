# Description

`vnodestats` is used to take statistics on a Moray database, which is part of the
Joyent Manta Project.

`vnodestats` currently returns the total number of vnodes (including duplicates),
the number of unique vnodes and the average distance between adjacent vnodes.

This data is used to show that a resharding correctly split the vnode-pnode
mapping in two. We expect the average distance between adjacent vnodes to double
in a simple resharding where the number of pnodes is doubled.

# Usage

```bash
$ # Copy vnodestats to a postgres zone
$ manta-login postgres
$ ./vnodestats
```

Build todo:

* In a SmartOS zone
```bash
    go build
```
* To cross compile (e.g. on macOS)
```bash
    GOOS=solaris GOARCH=amd64 go build 
```

## Connection configuration

The database connection is configured via enviroment variables.

* `DB_HOST` - defaults to localhost
* `DB_USER` - defaults to postgres
* `DB_PASSWORD` - defaults to postgres
* `DB_DATABASE` - defaults to moray

