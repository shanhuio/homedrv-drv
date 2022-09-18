package doorway

import (
	"context"
	"net/http"
	"net/url"

	"shanhu.io/aries/creds"
	"shanhu.io/misc/errcode"
)

// SimpleRouter provides a simple endpoint based router. It directly contacts
// the fabrics node for a token.
type SimpleRouter struct {
	Host    string // Host to route to.
	User    string
	Key     []byte
	KeyFile string

	Transport http.RoundTripper
}

// Route returns the host and the token.
func (r *SimpleRouter) Route(ctx context.Context) (string, string, error) {
	host := r.Host
	ep := &creds.Endpoint{
		Server:   (&url.URL{Scheme: "https", Host: host}).String(),
		User:     r.User,
		Key:      r.Key,
		PemFile:  r.KeyFile,
		Homeless: true,
		NoTTY:    true,
	}
	if r.Transport != nil {
		ep.Transport = r.Transport
	}
	login, err := creds.NewLogin(ep)
	if err != nil {
		return "", "", errcode.Annotate(err, "create login")
	}
	token, err := login.Token()
	if err != nil {
		return "", "", errcode.Annotate(err, "login")
	}
	return host, token, nil
}
