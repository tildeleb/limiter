Limiter
=======

This package implements an API rate limiter which can be used on either the client or server side. Limiter works to prevent live lock by waking waiting goroutines in FIFO order. The package is safe and can be used from goroutines without a wrapper mutex. It's written in Go.

Untested except for the example code.

