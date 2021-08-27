package doorway

import (
	"context"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/acme"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/netutil"
	"shanhu.io/misc/strutil"
	"shanhu.io/virgo/sniproxy"
)

// TLSProxyConfig is the configuration for the TLS proxy.
type TLSProxyConfig struct {
	// For these domains, forward the TCP connection directly.
	Forward map[string]string

	// Enables private mode. In private mode, unless listed
	// in PublicWhitelist, only connections for ACME ALPN challenges are
	// accepted.
	PrivateMode bool

	// Make these sites publicly accessible via fabrics.
	Public []string
}

type tlsProxy struct {
	lis tagConnListener

	forward map[string]string

	privateMode bool
	public      map[string]bool

	closing chan struct{}
}

func newTLSProxy(lis tagConnListener, config *TLSProxyConfig) *tlsProxy {
	p := &tlsProxy{
		lis:         lis,
		forward:     config.Forward,
		public:      strutil.MakeSet(config.Public),
		privateMode: config.PrivateMode,
		closing:     make(chan struct{}),
	}
	return p
}

func (p *tlsProxy) Addr() net.Addr { return p.lis.Addr() }

func isALPN(hello *sniproxy.TLSHelloInfo) bool {
	if hello.ServerName == "" {
		return false
	}
	return hello.ProtoCount == 1 && hello.FirstProto == acme.ALPNProto
}

func (p *tlsProxy) forwardTCPConn(conn *sniproxy.TLSHelloConn, addr string) {
	forward, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("dial %q for forwarding: %s", addr, err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
		case <-p.closing:
			cancel()
		}
	}()

	_ = netutil.JoinConn(ctx, conn, forward) // do not care about the error now.
}

func (p *tlsProxy) Close() error {
	close(p.closing)
	return p.lis.Close()
}

func (p *tlsProxy) Accept() (net.Conn, error) {
	for {
		conn, err := p.lis.acceptTag()
		if err != nil {
			return nil, err
		}

		h := sniproxy.NewTLSHelloConn(conn)
		hello, err := h.HelloInfo()
		if err != nil {
			// Drop connections that are not TLS.
			log.Println(errcode.Annotate(err, "init TLS connection"))
			h.Close()
			continue
		}
		name := strings.TrimSuffix(hello.ServerName, ".")

		if p.privateMode && conn.tag == tagFabrics {
			if _, found := p.public[name]; found {
				// pass
			} else if !isALPN(hello) {
				// Drop fabrics connections that are not ALPN
				log.Printf(
					"only alpn allowed from fabrics, got %q for %q",
					hello.FirstProto, hello.ServerName,
				)
				h.Close()
				continue
			}
		}

		if p.forward != nil {
			forward, ok := p.forward[name]
			if ok {
				go p.forwardTCPConn(h, forward)
				continue // ownership transferred
			}
		}

		return h, nil
	}
}
