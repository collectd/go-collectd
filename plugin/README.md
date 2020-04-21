## collectd plugins in Go

## About

This is _experimental_ code to write _collectd_ plugins in Go. That means the
API is not yet stable. It requires Go 1.13 or later and a recent version of the
collectd sources to build.

## Build

To set up your build environment, set the `CGO_CPPFLAGS` environment variable
so that _cgo_ can find the required header files:

    export COLLECTD_SRC="/path/to/collectd"
    export CGO_CPPFLAGS="-I${COLLECTD_SRC}/src/daemon -I${COLLECTD_SRC}/src"

You can then compile your Go plugins with:

    go build -buildmode=c-shared -o example.so

More information is available in the documentation of the `collectd.org/plugin`
package.

    godoc collectd.org/plugin

## Future

Only *log*, *read*, *write*, and *shutdown* callbacks are currently supported.
Based on these implementations it should be possible to implement the remaining
callbacks, even with little prior Cgo experience. The *init*, *flush*, and
*missing* callbacks are all likely low-hanging fruit. The *notification*
callback is a bit trickier because it requires implementing notifications in
the `collectd.org/api` package and the (un)marshaling of `notification_t`. The
(complex) *config* callback is currently work in progress, see #30.

If you're willing to give any of this a shot, please ping @octo to avoid
duplicate work.
