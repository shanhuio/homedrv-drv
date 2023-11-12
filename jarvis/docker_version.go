package jarvis

import (
	"fmt"
	"strconv"
	"strings"

	"shanhu.io/g/dock"
)

func checkDockerVersion(v *dock.VersionInfo) error {
	s := v.Version
	fields := strings.Split(s, ".")
	if len(fields) < 3 {
		return fmt.Errorf("invalid version string: %q", s)
	}

	major, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("invalid major in version %q", s)
	}
	minor, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("invalid minor in version %q", s)
	}

	num := fields[2]
	if pre, _, ok := strings.Cut(fields[2], "-"); ok {
		num = pre
	}
	maintanence, err := strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("invalid maintanence in version %q", s)
	}

	if major > 20 {
		return nil
	}
	if major == 20 && minor > 10 {
		return nil
	}
	if major == 20 && minor == 10 && maintanence >= 10 {
		return nil
	}

	return fmt.Errorf(
		"docker version %q too low, requires at least 20.10.10", s,
	)
}
