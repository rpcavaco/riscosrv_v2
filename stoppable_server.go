package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// baseado no código de valyala no issue #66 de fasthttp
type GracefulListener struct {
	// inner listener
	ln net.Listener

	// maximum wait time for graceful shutdown
	maxWaitTime time.Duration

	// this channel is closed during graceful shutdown on zero open connections.
	done chan struct{}

	// the number of open connections
	connsCount uint64

	// becomes non-zero when graceful shutdown starts
	shutdown uint64
}

// NewGracefulListener wraps the given listener into 'graceful shutdown' listener.
func newGracefulListener(ln net.Listener, maxWaitTime time.Duration) net.Listener {
	return &GracefulListener{
		ln:          ln,
		maxWaitTime: maxWaitTime,
		done:        make(chan struct{}),
	}
}

func (ln *GracefulListener) Accept() (net.Conn, error) {
	c, err := ln.ln.Accept()

	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&ln.connsCount, 1)

	return &gracefulConn{
		Conn: c,
		ln:   ln,
	}, nil
}

func (ln *GracefulListener) Addr() net.Addr {
	return ln.ln.Addr()
}

// Close closes the inner listener and waits until all the pending open connections
// are closed before returning.
func (ln *GracefulListener) Close() error {
	err := ln.ln.Close()

	if err != nil {
		return nil
	}

	return ln.waitForZeroConns()
}

func (ln *GracefulListener) waitForZeroConns() error {
	atomic.AddUint64(&ln.shutdown, 1)

	if atomic.LoadUint64(&ln.connsCount) == 0 {
		close(ln.done)
		return nil
	}

	select {
	case <-ln.done:
		return nil
	case <-time.After(ln.maxWaitTime):
		return fmt.Errorf("cannot complete graceful shutdown in %s", ln.maxWaitTime)
	}

}

func (ln *GracefulListener) closeConn() {
	connsCount := atomic.AddUint64(&ln.connsCount, ^uint64(0))

	if atomic.LoadUint64(&ln.shutdown) != 0 && connsCount == 0 {
		close(ln.done)
	}
}

type gracefulConn struct {
	net.Conn
	ln *GracefulListener
}

func (c *gracefulConn) Close() error {
	err := c.Conn.Close()

	if err != nil {
		return err
	}

	c.ln.closeConn()

	return nil
}

func doServe(s *fasthttp.Server, ln net.Listener, purpose string) {
	LogInfof("stopping fasthttp server (%s), out: %v", purpose, s.Serve(ln))
}

// Cria um servidor HTTP na endereço indicado em addr.
// Devolve um net.Listener ao qual se pode aplicar o método Close para terminar o servidor.
//   - purpose: HTTP, WebSocket (string informativa do objectivo do servidor)
func AsyncListenAndServe(addr string, handler fasthttp.RequestHandler, maxWaitTime time.Duration, purpose string) (net.Listener, error) {
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}
	gln := newGracefulListener(ln, maxWaitTime)
	s := &fasthttp.Server{
		Handler: handler,
	}
	LogInfof("starting fasthttp server on %s (%s)", addr, purpose)
	go doServe(s, gln, purpose)

	return gln, err
}

/*func CompressHandlerLevel(s *appServer, h fasthttp.RequestHandler) RequestHandler2 {
	return func(s *appServer, ctx *RequestCtx) {
		h(s, ctx)
		ce := ctx.Response.Header.PeekBytes(strContentEncoding)
		if len(ce) > 0 {
			// Do not compress responses with non-empty
			// Content-Encoding.
			return
		}
		if ctx.Request.Header.HasAcceptEncodingBytes(fasthttp.strGzip) {
			ctx.Response.gzipBody(fasthttp.CompressDefaultCompression)
		} else if ctx.Request.Header.HasAcceptEncodingBytes(fasthttp.strDeflate) {
			ctx.Response.deflateBody(fasthttp.CompressDefaultCompression)
		}
	}
}*/
