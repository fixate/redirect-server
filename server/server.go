package server

import (
	"net/http"
	"time"

	mfst "github.com/fixate/redirect-server/manifest"
)

type ServerOptions struct {
	Manifest *mfst.Manifest
	Bind     string
}

func NewServer(options *ServerOptions) *http.Server {
	return &http.Server{
		Addr:           options.Bind,
		Handler:        newHandler(options.Manifest),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
