# check_mysql_stats

A simple tool to check some MySQL/MariaDB statistics gathered from 
`'information_schema.global_stats'`.
It's not pretty but I wanted to play with Go a bit and while doing so create 
something I could use as well.

## Build

- Install Go
- Clone the repostory
- run `go build`

## Parameters

* `-h` - MySQL/MariaDB host to connect to (default: 127.0.0.1)
* `-p` - the port MySQL/MariaDB listens on (default: 3306)
* `-u` - the username used for connecting to MySQL/MariaDB
* `-p` - the password to use when connecting
* `-rwarn` - warning level for read requests (in ops/s, default: 250)
* `-rcrit` - critical level for read requests (in ops/s, default: 500)
* `-wwarn` - warning level for write requests (in ops/s, default: 50)
* `-wcrit` - critical level for write requests (in ops/s, default: 100)
* `-t` - timeout when connecting to MySQL/MariaDB (in s, default: 5)
* `-v` - display version information
* `-help` - display help/usage information

## Output
The plugin will output a string of the format

`<state>: <reason> | <metrics>`

### state
can be on of the following:
* `OK` - everything is fine
* `WARNING`- at least one of either the read or the write ops warning levels has 
been reached
* `CRITICAL`- at least on of either the read or the write ops warning level has 
been reached

### reason
Just a text describing the state of the MySQL/MaraiDB instance.

### metrics
Multiple values separated by spaces giving some statistics about the 
MySQL/MariaDB instance:

* `queries` - total number of queries since last restart
* `selects` - total number of SELECT queries since last restart
* `inserts` - total number of INSERT queries since last restart
* `updates` - total number of UPDATE queries since last restart
* `deletes` - total number of DELETE queries since last restart
* `uptime` - instance uptime in seconds
* `reads_per_second` - calculated number of read requests per second since last 
run (or last restart if first run)
* `writes_per_second` - calculated number of write requests per second since 
last run (or last restart if first run)

### Data Store
For every host checked the tool will create a JSON file containing the 
measurements taken on the last run. The file is placed into the same directory 
where the tool resides with a filename of

`<name of tool>.<host address as passed by icinga>.stats`

These files can be safely deleted. In case they don't exist a new version will
be created and the number of seconds since the last restart as reported by 
MySQL/MariaDB will be used to generate the Rps/Wps counters. 
