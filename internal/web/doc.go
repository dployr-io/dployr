// Package web provides the HTTP server and multiplexer for the daemon's web API.
// It wires together core handlers (deployment, service, proxy, log streaming, system)
// and manages all external HTTP endpoints exposed by the daemon.
// Concrete HTTP routing and server lifecycle logic are contained in this package.
package web
