## collectd plugins in Go

This is _extremely experimental_ code to write _collectd_ plugins in Go. It
requires Go 1.5 and a recent version of the collectd sources to build.

To set up your build environment, set the `CGO_CPPFLAGS` environment variable
so that _cgo_ can find the required header files:

    export COLLECTD_SRC="/path/to/collectd"
    export CGO_CPPFLAGS="-I${COLLECTD_SRC}/src/daemon -I${COLLECTD_SRC}/src"

You can then compile your Go plugins with:

    go build -buildmode=c-shared -o example.so

More information is available in the documentation of the "collectd.org/plugin"
package.
