Limiter
=======

This package implements an API rate limiter which can be used on either the client or server side. The rate limiting is specified as calls per duration. Limiter works to prevent live lock by waking waiting goroutines in FIFO order. The package is safe and can be used from goroutines *without* a wrapper mutex. Written in Go.

Completely untested except for the example code.

