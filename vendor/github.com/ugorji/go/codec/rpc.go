// Copyright (c) 2012-2018 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package codec

import (
	"bufio"
	"errors"
	"io"
	"net/rpc"
	"sync"
)

// Rpc provides a rpc Server or Client Codec for rpc communication.
type Rpc interface {
	ServerCodec(conn io.ReadWriteCloser, h Handle) rpc.ServerCodec
	ClientCodec(conn io.ReadWriteCloser, h Handle) rpc.ClientCodec
}

// RPCOptions holds options specific to rpc functionality
type RPCOptions struct {
	// RPCNoBuffer configures whether we attempt to buffer reads and writes during RPC calls.
	//
	// Set RPCNoBuffer=true to turn buffering off.
	// Buffering can still be done if buffered connections are passed in, or
	// buffering is configured on the handle.
	RPCNoBuffer bool
}

// rpcCodec defines the struct members and common methods.
type rpcCodec struct {
	c io.Closer
	r io.Reader
	w io.Writer
	f ioFlusher

	dec *Decoder
	enc *Encoder
	// bw  *bufio.Writer
	// br  *bufio.Reader
	mu sync.Mutex
	h  Handle

	cls    bool
	clsmu  sync.RWMutex
	clsErr error
}

func newRPCCodec(conn io.ReadWriteCloser, h Handle) rpcCodec {
	// return newRPCCodec2(bufio.NewReader(conn), bufio.NewWriter(conn), conn, h)
	return newRPCCodec2(conn, conn, conn, h)
}

func newRPCCodec2(r io.Reader, w io.Writer, c io.Closer, h Handle) rpcCodec {
	// defensive: ensure that jsonH has TermWhitespace turned on.
	if jsonH, ok := h.(*JsonHandle); ok && !jsonH.TermWhitespace {
		panic(errors.New("rpc requires a JsonHandle with TermWhitespace set to true"))
	}
	// always ensure that we use a flusher, and always flush what was written to the connection.
	// we lose nothing by using a buffered writer internally.
	f, ok := w.(ioFlusher)
	bh := h.getBasicHandle()
	if !bh.RPCNoBuffer {
		if bh.WriterBufferSize <= 0 {
			if !ok {
				bw := bufio.NewWriter(w)
				f, w = bw, bw
			}
		}
		if bh.ReaderBufferSize <= 0 {
			if _, ok = w.(ioPeeker); !ok {
				if _, ok = w.(ioBuffered); !ok {
					br := bufio.NewReader(r)
					r = br
				}
			}
		}
	}
	return rpcCodec{
		c:   c,
		w:   w,
		r:   r,
		f:   f,
		h:   h,
		enc: NewEncoder(w, h),
		dec: NewDecoder(r, h),
	}
}

func (c *rpcCodec) write(obj1, obj2 interface{}, writeObj2 bool) (err error) {
	if c.isClosed() {
		return c.clsErr
	}
	err = c.enc.Encode(obj1)
	if err == nil {
		if writeObj2 {
			err = c.enc.Encode(obj2)
		}
		// if err == nil && c.f != nil {
		// 	err = c.f.Flush()
		// }
	}
	if c.f != nil {
		if err == nil {
			err = c.f.Flush()
		} else {
			c.f.Flush()
		}
	}
	return
}

func (c *rpcCodec) swallow(err *error) {
	defer panicToErr(c.dec, err)
	c.dec.swallow()
}

func (c *rpcCodec) read(obj interface{}) (err error) {
	if c.isClosed() {
		return c.clsErr
	}
	//If nil is passed in, we should read and discard
	if obj == nil {
		// var obj2 interface{}
		// return c.dec.Decode(&obj2)
		c.swallow(&err)
		return
	}
	return c.dec.Decode(obj)
}

func (c *rpcCodec) isClosed() (b bool) {
	if c.c != nil {
		c.clsmu.RLock()
		b = c.cls
		c.clsmu.RUnlock()
	}
	return
}

func (c *rpcCodec) Close() error {
	if c.c == nil || c.isClosed() {
		return c.clsErr
	}
	c.clsmu.Lock()
	c.cls = true
	// var fErr error
	// if c.f != nil {
	// 	fErr = c.f.Flush()
	// }
	// _ = fErr
	// c.clsErr = c.c.Close()
	// if c.clsErr == nil && fErr != nil {
	// 	c.clsErr = fErr
	// }
	c.clsErr = c.c.Close()
	c.clsmu.Unlock()
	return c.clsErr
}

func (c *rpcCodec) ReadResponseBody(body interface{}) error {
	return c.read(body)
}

// -------------------------------------

type goRpcCodec struct {
	rpcCodec
}

func (c *goRpcCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	// Must protect for concurrent access as per API
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.write(r, body, true)
}

func (c *goRpcCodec) WriteResponse(r *rpc.Response, body interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.write(r, body, true)
}

func (c *goRpcCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.read(r)
}

func (c *goRpcCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.read(r)
}

func (c *goRpcCodec) ReadRequestBody(body interface{}) error {
	return c.read(body)
}

// -------------------------------------

// goRpc is the implementation of Rpc that uses the communication protocol
// as defined in net/rpc package.
type goRpc struct{}

// GoRpc implements Rpc using the communication protocol defined in net/rpc package.
//
// Note: network connection (from net.Dial, of type io.ReadWriteCloser) is not buffered.
//
// For performance, you should configure WriterBufferSize and ReaderBufferSize on the handle.
// This ensures we use an adequate buffer during reading and writing.
// If not configured, we will internally initialize and use a buffer during reads and writes.
// This can be turned off via the RPCNoBuffer option on the Handle.
//   var handle codec.JsonHandle
//   handle.RPCNoBuffer = true // turns off attempt by rpc module to initialize a buffer
//
// Example 1: one way of configuring buffering explicitly:
//   var handle codec.JsonHandle // codec handle
//   handle.ReaderBufferSize = 1024
//   handle.WriterBufferSize = 1024
//   var conn io.ReadWriteCloser // connection got from a socket
//   var serverCodec = GoRpc.ServerCodec(conn, handle)
//   var clientCodec = GoRpc.ClientCodec(conn, handle)
//
// Example 2: you can also explicitly create a buffered connection yourself,
// and not worry about configuring the buffer sizes in the Handle.
//   var handle codec.Handle     // codec handle
//   var conn io.ReadWriteCloser // connection got from a socket
//   var bufconn = struct {      // bufconn here is a buffered io.ReadWriteCloser
//       io.Closer
//       *bufio.Reader
//       *bufio.Writer
//   }{conn, bufio.NewReader(conn), bufio.NewWriter(conn)}
//   var serverCodec = GoRpc.ServerCodec(bufconn, handle)
//   var clientCodec = GoRpc.ClientCodec(bufconn, handle)
//
var GoRpc goRpc

func (x goRpc) ServerCodec(conn io.ReadWriteCloser, h Handle) rpc.ServerCodec {
	return &goRpcCodec{newRPCCodec(conn, h)}
}

func (x goRpc) ClientCodec(conn io.ReadWriteCloser, h Handle) rpc.ClientCodec {
	return &goRpcCodec{newRPCCodec(conn, h)}
}
