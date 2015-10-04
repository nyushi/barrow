package barrow

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/flynn/go-shlex"

	"gopkg.in/yaml.v2"
)

type script struct {
	Cmd  string
	Once bool
}

func (s *script) Run(dryrun bool) error {
	if !dryrun {
		args, err := shlex.Split(s.Cmd)
		if err != nil {
			return fmt.Errorf("failed to parse command line `%s`: %s", s.Cmd, err)
		}
		c := exec.Command(args[0], args[1:]...)
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute command: `%+v`: %s", args, err)
		}
		outstr := string(out)
		if len(outstr) > 0 {
			fmt.Println(outstr)
		}
	}
	return nil
}

type Hooks struct {
	Scripts map[string]*script  `yaml:"scripts"`  // key:script name
	AtFirst []string            `yaml:"at_first"` // slice of script name
	Changed map[string][]string `yaml:"changed"`  // key:filename, val:script name
	AtLast  []string            `yaml:"at_last"`  // slice of script name
}

func ParseHooks(path string) (*Hooks, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var h Hooks
	if err := yaml.Unmarshal(b, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (h *Hooks) Validate() (string, error) {
	// TODO
	return "", nil
}
