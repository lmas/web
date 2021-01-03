package web

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

// Security Sources:
// https://blog.cloudflare.com/exposing-go-on-the-internet/ (out of date)
// https://wiki.mozilla.org/Security/Server_Side_TLS
// https://go.googlesource.com/go/+blame/go1.15.6/src/crypto/tls/common.go

type ServerOptions struct {
	Addr    string
	Handler http.Handler
	Log     *log.Logger
}

func NewServer(opt ServerOptions) *http.Server {
	tlsConf := &tls.Config{
		MinVersion: tls.VersionTLS13,
		// Don't need to change these, as go1.15 has pretty good defaults
		// (as per mozilla's recommendations, for modern config)
		// CipherSuites:
		// CurvePreferences:
		//PreferServerCipherSuites: true,

		// TODO: might need some more options when enabling auto tls, see:
		// https://github.com/golang/crypto/blob/eec23a3978ad/acme/autocert/autocert.go#L220
	}

	// TODO: start a 2nd http server that redirects to https, see the cloudflare
	// blog and mozilla tls generator for examples

	// NOTE: no need to set a tcp keep-alive, go1.15 has a default of 15s, see
	// https://go.googlesource.com/go/+/go1.15.6/src/net/dial.go#17
	// https://github.com/golang/go/issues/31510

	return &http.Server{
		Addr:        opt.Addr,
		Handler:     opt.Handler,
		ErrorLog:    opt.Log,
		TLSConfig:   tlsConf,
		ReadTimeout: 10 * time.Second,
		// TODO: might to want ReadHeaderTimeout instead, so handlers can
		// decide for themselves when to time out a body read?
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		// MaxHeaderBytes: defaults to 1mb, per http.DefaultMaxHeaderBytes
	}
}
