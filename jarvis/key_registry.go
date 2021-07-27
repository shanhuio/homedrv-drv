package jarvis

import (
	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/rsautil"
)

type keyRegistry struct {
	users *users
}

func newKeyRegistry(users *users) *keyRegistry {
	return &keyRegistry{users: users}
}

func (r *keyRegistry) set(user string, keyBytes []byte) error {
	if user != rootUser {
		return errcode.InvalidArgf("only root user supported")
	}
	return r.users.mutate(user, func(info *userInfo) error {
		info.APIKeys = keyBytes
		return nil
	})
}

func (r *keyRegistry) Keys(user string) ([]*rsautil.PublicKey, error) {
	if user != rootUser {
		return nil, errcode.InvalidArgf("only root user supported")
	}

	info, err := r.users.get(user)
	if err != nil {
		return nil, errcode.Annotate(err, "get user info")
	}
	return rsautil.ParsePublicKeys(info.APIKeys)
}

func (r *keyRegistry) apiSet(c *aries.C, keyBytes []byte) error {
	return r.set(rootUser, keyBytes)
}
