nanny
=====

`nanny` monitors changes to a file or directory (recursively) and runs some command(s) when a change is detected.

`nanny` is a bit like [atchange](http://www.ccrnp.ncifcrf.gov/~toms/atchange.html) but written in [Go](http://golang.org).

Note: not tested and unlikely to work on Windows and (but shouldn't require big changes).

Usage
-----

    nanny path/to/file/or/dir commands

Whenever a change is detected, the `commands` are run in a new `$SHELL` sub-process and their `stdout` and `stderr` are redirected to the console.

Examples
--------

1. Re-build a program whenever change is detected in the current directory:

        nanny . "go build nanny.go"

    Watch a file and when it changes echo current date, re-compile, echo "OK" or "ERROR" depending on compile command exit code:

    	nanny vhost.go "date; (go build vhost.go && echo OK) || echo ERROR"

2. Generate markup and open in browser:

        nanny README.markdown "markdown README.markdown > /tmp/temp.html; open /tmp/temp.html"

Installation
------------

Install [Go](http://golang.org), then

    go get github.org/maurice/nanny


