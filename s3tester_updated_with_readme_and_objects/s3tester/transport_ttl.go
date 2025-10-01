package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type connTracker struct {
	ttl   time.Duration
	mu    sync.Mutex
	conns map[*TimedConn]struct{}
}

func newConnTracker(ttl time.Duration) *connTracker {
	if ttl <= 0 {
		return nil
	}
	return &connTracker{
		ttl:   ttl,
		conns: make(map[*TimedConn]struct{}),
	}
}

func (ct *connTracker) add(conn *TimedConn) {
	ct.mu.Lock()
	ct.conns[conn] = struct{}{}
	ct.mu.Unlock()
}

func (ct *connTracker) remove(conn *TimedConn) {
	if ct == nil {
		return
	}
	ct.mu.Lock()
	delete(ct.conns, conn)
	ct.mu.Unlock()
}

func (ct *connTracker) reap() {
	if ct == nil {
		return
	}
	cutoff := ct.ttl
	var expired []*TimedConn
	ct.mu.Lock()
	for conn := range ct.conns {
		if conn.Age() >= cutoff && !conn.InUse() {
			expired = append(expired, conn)
			delete(ct.conns, conn)
		}
	}
	ct.mu.Unlock()
	for _, conn := range expired {
		_ = conn.Close()
	}
}

// TimedConn wraps a net.Conn and records its creation time.
type TimedConn struct {
	net.Conn
	tracker  *connTracker
	created  int64 // unix nano
	inFlight int32
	closed   int32
}

func NewTimedConn(tracker *connTracker, c net.Conn) *TimedConn {
	tc := &TimedConn{
		Conn:    c,
		tracker: tracker,
		created: time.Now().UnixNano(),
	}
	if tracker != nil {
		tracker.add(tc)
	}
	return tc
}

func (t *TimedConn) Age() time.Duration {
	return time.Since(time.Unix(0, atomic.LoadInt64(&t.created)))
}

func (t *TimedConn) InUse() bool {
	return atomic.LoadInt32(&t.inFlight) > 0
}

func (t *TimedConn) Read(b []byte) (int, error) {
	atomic.AddInt32(&t.inFlight, 1)
	defer atomic.AddInt32(&t.inFlight, -1)
	return t.Conn.Read(b)
}

func (t *TimedConn) Write(b []byte) (int, error) {
	atomic.AddInt32(&t.inFlight, 1)
	defer atomic.AddInt32(&t.inFlight, -1)
	return t.Conn.Write(b)
}

func (t *TimedConn) Close() error {
	if atomic.CompareAndSwapInt32(&t.closed, 0, 1) {
		if t.tracker != nil {
			t.tracker.remove(t)
		}
		return t.Conn.Close()
	}
	return nil
}

// StartConnTTLReaper runs a background goroutine that closes idle connections
// when they age beyond connTTL. It only reaps connections that are not currently in use.
func StartConnTTLReaper(tr *http.Transport, tracker *connTracker, connTTL time.Duration, interval time.Duration, ctx context.Context) context.CancelFunc {
	if tracker == nil || connTTL <= 0 {
		return func() {}
	}
	if interval <= 0 {
		interval = connTTL / 2
		if interval < time.Second {
			interval = time.Second
		}
	}
	rctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-rctx.Done():
				return
			case <-ticker.C:
				tracker.reap()
				tr.CloseIdleConnections()
			}
		}
	}()
	return cancel
}
