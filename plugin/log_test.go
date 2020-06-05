package plugin_test

import (
	"errors"
	"log"
	"net/http"

	"collectd.org/plugin"
)

func ExampleLogWriter() {
	l := log.New(plugin.LogWriter(plugin.SeverityError), "", log.Lshortfile)

	// Start an HTTP server that logs errors to collectd's logging facility.
	srv := &http.Server{
		ErrorLog: l,
	}
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		l.Println("ListenAndServe:", err)
	}
}
