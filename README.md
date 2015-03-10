# go-collectd

Utilities for using [collectd](https://collectd.org/) together with [Go](http://golang.org/).

# Synopsis

    import (
        "time"
        
        "collectd.org/api"
        "collectd.org/exec"
    )
    
    vl := ValueList{
       Identifier{
           Host: exec.Hostname(),
           Plugin: "golang",
           Type: "gauge",
       },
       Time: time.Now(),
       Interval: exec.Interval(),
       Values: []api.Value{api.Gauge(42)},
    }
    exec.Dispatch(vl)

# Description

This is a very simple package and very much a *Work in Progress*, so expect
things to move around and be renamed a lot.

The reposiroty is organized as follows:

* Package `collectd.org/api` declares data structures you may already know from
  the *collectd* source code itself, such as `ValueList`.
* Package `collectd.org/format` declares functions for formatting *ValueLists*
  in other format. Right now, only `PUTVAL` is implemented. Eventually I plan
  to add parsers for some formats, such as the JSON export. A converter to/from
  the binary network protocol might also go here.
* Package `collectd.org/exec` declares some utilities for writing binaries to
  be executed with the *exec plugin*. It provides some utilities (getting the
  hostname, e.g.) and an executor which you may use to easily schedule function
  calls.

# Author

Florian "octo" Forster <ff at octo.it>
