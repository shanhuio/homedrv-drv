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
	"fmt"
	"net"
	"sync"
)

type tunnelListener struct {
	tcp       net.Listener
	tunnel    net.Listener
	ch        chan net.Conn
	errCh     chan error
	closed    chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func (lis *tunnelListener) Accept() (net.Conn, error) {
	select {
	case c := <-lis.ch:
		return c, nil
	case err := <-lis.errCh:
		return nil, err
	case <-lis.closed:
		return nil, fmt.Errorf("connection closed")
	}
}

func (lis *tunnelListener) bg(l net.Listener) {
	defer lis.wg.Done()
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-lis.closed:
				return
			case lis.errCh <- err:
				return
			}
		} else {
			select {
			case <-lis.closed:
				return
			case lis.ch <- conn:
			}
		}
	}
}

func (lis *tunnelListener) doClose() error {
	tcpErr := lis.tcp.Close()
	tunnelErr := lis.tunnel.Close()
	close(lis.closed)
	lis.wg.Wait()
	if tcpErr != nil {
		return tcpErr
	}
	return tunnelErr
}

func (lis *tunnelListener) Close() error {
	err := errors.New("already closed")
	lis.closeOnce.Do(func() {
		err = lis.doClose()
	})
	return err
}

func (lis *tunnelListener) Addr() net.Addr {
	return &tunnelAddr{
		addr: fmt.Sprintf(
			"tcp:%s+tunnel:%s",
			lis.tcp.Addr(), lis.tunnel.Addr(),
		),
	}
}

type tunnelAddr struct{ addr string }

func (a *tunnelAddr) Network() string { return "tcptunn" }
func (a *tunnelAddr) String() string  { return a.addr }

func newTunnelListener(tcp, tunnel net.Listener) *tunnelListener {
	lis := &tunnelListener{
		tcp:    tcp,
		tunnel: tunnel,
		ch:     make(chan net.Conn),
		errCh:  make(chan error),
		closed: make(chan struct{}),
	}
	lis.wg.Add(2)
	go lis.bg(tcp)
	go lis.bg(tunnel)

	return lis
}
