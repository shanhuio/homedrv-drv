// Copyright (C) 2021  Shanhu Tech Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package doorway

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type reconnectListener struct {
	listen        func() (net.Listener, error)
	errorCallback func(err error)

	mu   sync.Mutex
	addr net.Addr
	conn chan net.Conn

	closeOnce sync.Once
	closed    chan struct{}
}

func newReconnectListener(
	listen func() (net.Listener, error),
	onError func(err error),
) (*reconnectListener, error) {
	first, err := listen()
	if err != nil {
		return nil, err
	}

	lis := &reconnectListener{
		listen:        listen,
		errorCallback: onError,
		addr:          first.Addr(),
		conn:          make(chan net.Conn),
		closed:        make(chan struct{}),
	}
	go lis.bg(first)
	return lis, nil
}

func (l *reconnectListener) setAddr(addr net.Addr) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.addr = addr
}

func (l *reconnectListener) isClosed() bool {
	select {
	case <-l.closed:
		return true
	default:
		return false
	}
}

func (l *reconnectListener) Close() error {
	err := errors.New("already closed")
	l.closeOnce.Do(func() {
		close(l.closed)
		err = nil
	})
	return err
}

func (l *reconnectListener) callback(err error) {
	if l.errorCallback == nil {
		return
	}
	l.errorCallback(err)
}

func (l *reconnectListener) Addr() net.Addr {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.addr == nil {
		return emptyAddr{}
	}
	return l.addr
}

type emptyAddr struct{}

func (emptyAddr) Network() string { return "empty" }
func (emptyAddr) String() string  { return "empty" }

func (l *reconnectListener) sleep(dur time.Duration) {
	timer := time.NewTimer(dur)
	defer timer.Stop()
	select {
	case <-l.closed:
		return
	case <-timer.C:
		return
	}
}

func (l *reconnectListener) bgAccept(lis net.Listener) {
	defer log.Println("listener lost")
	defer lis.Close()

	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-l.closed:
			lis.Close() // to unblock accept.
		}
	}()
	defer close(done)

	for {
		if l.isClosed() {
			return
		}
		conn, err := lis.Accept()
		if err != nil {
			l.callback(err)
			return
		}

		select {
		case l.conn <- conn:
		case <-l.closed:
			return
		}
	}
}

func (l *reconnectListener) bg(lis net.Listener) {
	if lis != nil {
		l.bgAccept(lis)
		l.sleep(time.Second)
	}

	for !l.isClosed() {
		lis, err := l.listen()
		if err != nil {
			l.callback(err)
			l.sleep(time.Second)
			continue
		}
		addr := lis.Addr()
		log.Printf("reconnected: %s", addr)
		l.setAddr(addr)

		l.bgAccept(lis)
		l.sleep(time.Second)
	}
}

func (l *reconnectListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conn:
		return conn, nil
	case <-l.closed:
		return nil, io.EOF
	}
}
