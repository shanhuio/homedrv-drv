package homerelease

import (
	"os"
	"sort"

	"shanhu.io/misc/errcode"
	"shanhu.io/misc/tarutil"
)

func writeObjects(p string, objects map[string]string) error {
	var sums []string
	for sum := range objects {
		sums = append(sums, sum)
	}
	sort.Strings(sums)

	stream := tarutil.NewStream()
	for _, sum := range sums {
		stream.AddFile(sum, tarutil.ModeMeta(0644), objects[sum])
	}

	f, err := os.Create(p)
	if err != nil {
		return errcode.Annotate(err, "create file")
	}
	defer f.Close()

	if _, err := stream.WriteTo(f); err != nil {
		return errcode.Annotate(err, "write tarball")
	}
	if err := f.Sync(); err != nil {
		return errcode.Annotate(err, "sync to disk")
	}
	return nil
}
