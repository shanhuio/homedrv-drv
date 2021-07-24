package homerelease

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"shanhu.io/misc/errcode"
)

type dockerImage struct {
	id  string
	sum string
}

type dockerManifest struct {
	Config string `json:",omitempty"`
}

func extractDockerManifest(r io.Reader) ([]*dockerManifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, errcode.Annotate(err, "new gzip reader")
	}
	t := tar.NewReader(gz)
	var m []*dockerManifest
	for {
		h, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errcode.Annotate(err, "read tarball")
		}

		if h.Typeflag == tar.TypeReg && h.Name == "manifest.json" {
			bs, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, errcode.Annotate(err, "read manifest")
			}
			if err := json.Unmarshal(bs, &m); err != nil {
				return nil, errcode.Annotate(err, "unmarshal manifest")
			}

			return m, nil
		}
	}

	return nil, errcode.NotFoundf("manifest not found")
}

func sumDockerTgz(p string) (*dockerImage, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errcode.Annotate(err, "open docker image file")
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, errcode.Annotate(err, "calculate checksum")
	}
	sum := h.Sum(nil)
	sumStr := "sha256:" + hex.EncodeToString(sum[:])

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, errcode.Annotate(err, "reset to the file start")
	}
	m, err := extractDockerManifest(f)
	if err != nil {
		return nil, errcode.Annotate(err, "extract docker manifest")
	}
	if len(m) != 1 {
		return nil, errcode.InvalidArgf("contains %d images, not 1", len(m))
	}
	id := strings.TrimSuffix(m[0].Config, ".json")
	if id == "" {
		return nil, errcode.InvalidArgf("empty docker id")
	}
	return &dockerImage{
		id:  id,
		sum: sumStr,
	}, nil
}
