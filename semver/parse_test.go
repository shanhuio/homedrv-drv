package semver

import (
	"testing"
)

func TestMajor(t *testing.T) {
	for _, test := range []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "1.0.0", want: 1},
		{input: "30.4.3", want: 30},
		{input: "24.12.343.7", want: 24},
		{input: "1", want: 1},
		{input: "3.7", want: 3},
		{input: "", wantErr: true},
		{input: "a.b", wantErr: true},
		{input: "..........", wantErr: true},
	} {
		got, err := Major(test.input)
		if err != nil {
			if !test.wantErr {
				t.Errorf("Major(%q), got error: %s", test.input, err)
			}
			continue
		}

		if test.wantErr {
			t.Errorf("Major(%q), got %d, want error", test.input, got)
		} else if got != test.want {
			t.Errorf("Major(%q), got %d, want %d", test.input, got, test.want)
		}
	}
}
