package jarvis

import (
	"testing"

	"shanhu.io/g/dock"
)

func TestCheckDockerVersion(t *testing.T) {
	for _, v := range []string{
		"20.10.10",
		"20.10.10-debug",
		"20.10.10-rc",
		"20.10.22",
		"20.11.22",
		"21.0.0",
		"100.0.0",
	} {
		info := &dock.VersionInfo{Version: v}
		if err := checkDockerVersion(info); err != nil {
			t.Errorf("%q should not error, got %v", v, err)
		}
	}

	for _, v := range []string{
		"",
		"1",
		"a",
		"1.0",
		"19.10.9",
		"20.10.9-rc",
		"20.10.9-debug",
		"20.10.9",
	} {
		info := &dock.VersionInfo{Version: v}
		if err := checkDockerVersion(info); err == nil {
			t.Errorf("%q should error, got nil", v)
		}
	}

}
