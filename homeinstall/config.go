// Copyright (C) 2023  Shanhu Tech Inc.
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

package homeinstall

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"shanhu.io/g/errcode"
)

type installScript struct {
	endpoint string
	passCode string
	bin      string
}

func (s *installScript) args() []string {
	return []string{
		"sudo", s.bin,
		"--name", s.endpoint,
		"--code", s.passCode,
	}
}

func (s *installScript) bashScript() []byte {
	buf := new(bytes.Buffer)
	io.WriteString(buf, "#/bin/bash\n")
	io.WriteString(buf, strings.Join(s.args(), " "))
	io.WriteString(buf, "\n")
	return buf.Bytes()
}

func (s *installScript) writeOut(f string) error {
	bs := s.bashScript()
	if f == "" {
		_, err := os.Stdout.Write(bs)
		return err
	}
	return ioutil.WriteFile(f, bs, 0755)
}

func checkEndpointName(name string) error {
	if name == "" {
		return errcode.InvalidArgf("endpoint name is empty")
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' {
			continue
		}
		return errcode.InvalidArgf("invalid endpoint name %q", name)
	}
	return nil
}

func checkPassCode(code string) error {
	if code == "" {
		return errcode.InvalidArgf("one-time code is empty")
	}

	for _, r := range code {
		if r >= '0' && r <= '9' {
			continue
		}
		return errcode.InvalidArgf("invalid one-time code: %q", code)
	}
	return nil
}

func cmdConfig(args []string) error {
	flags := cmdFlags.New()
	bin := flags.String(
		"bin", "/opt/homedrv/install",
		"install program to run",
	)
	output := flags.String("out", "", "output file to write into")

	flags.Parse(args)

	in := bufio.NewReader(os.Stdin)

	fmt.Println(
		"Please visit https://www.homedrive.io/endpoints " +
			"and select an endpoint.",
	)
	fmt.Println("You can also create an new one.")

	fmt.Print("Please input the endpoint name: ")
	endpoint, err := in.ReadString('\n')
	if err != nil {
		return errcode.Annotate(err, "read endpoint name")
	}

	endpoint = strings.TrimSpace(endpoint)
	if err := checkEndpointName(endpoint); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Now, please create a new one-time installation code at:")
	fmt.Printf("https://www.homedrive.io/endpoint/%s\n", endpoint)
	fmt.Print("And input the one-time installation code here: ")

	passCode, err := in.ReadString('\n')
	if err != nil {
		return errcode.Annotate(err, "read one-time code")
	}
	passCode = strings.TrimSpace(passCode)
	if err := checkPassCode(passCode); err != nil {
		return err
	}

	fmt.Println()

	s := &installScript{
		endpoint: endpoint,
		passCode: passCode,
		bin:      *bin,
	}
	return s.writeOut(*output)
}
