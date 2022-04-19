/*
Copyright © 2022 Optriment
*/
package cmd

import (
	"bytes"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getGIDstring() string {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b)
}

type goroutineID struct {
	Name string
}

func (gid *goroutineID) String() string {
	return getGIDstring()
}

func (gid *goroutineID) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str(gid.Name, getGIDstring())
}

// startPprof starts debug service
// just follow /debug/pprof/
func startPprof(hostport string) {
	log.Debug().Msgf("Starting pprof http server at %s", hostport)
	go func() {
		log.Info().
			Err(http.ListenAndServe(hostport, nil)).Send()
	}()
}
