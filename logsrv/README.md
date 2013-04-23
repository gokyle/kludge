logsrv

`logsrv` is the central logging server for the kludge system. Systems
can use the `logsrv` package in the `logsrv` subdirectory to set up
logging to the server.

It may be installed with `make install`, which sets up the upstart service.

By default, it listens on port 5988, and logs to the 'logs.db' SQlite3
database. These behaviours may be changed with `-p` to set the port,
and `-f` to change the database file.
