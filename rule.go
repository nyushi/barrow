package barrow

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/nyushi/install"
)

type Rule struct {
	Dir      string
	FileList FileList
	Hooks    *Hooks
}

func LoadRule(path string) (*Rule, error) {
	r := &Rule{Dir: path}

	flPath := fmt.Sprintf("%s/FILES", path)
	fl, err := ParseFileList(flPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %s", flPath, err)
	}
	for _, fi := range fl {
		fi.SrcPath = r.srcPath(fi)
	}

	r.FileList = fl

	hookPath := fmt.Sprintf("%s/HOOKS", path)
	r.Hooks, err = ParseHooks(hookPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %s", hookPath, err)
	}

	if err := r.validate(); err != nil {
		return nil, fmt.Errorf("validation error: %s", err)
	}

	return r, nil
}
func (r *Rule) rootPath() string {
	return path.Join(r.Dir, "ROOT")
}
func (r *Rule) srcPath(fi *FileInfo) string {
	return path.Join(r.rootPath(), fi.Path)
}
func (r *Rule) validate() error {
	msgs := []string{}

	for _, f := range []func() (string, error){
		r.validateUsers,
		r.validateGroups,
		r.validateROOT,
		r.Hooks.Validate,
	} {
		msg, err := f()
		if err != nil {
			return err
		}
		if msg != "" {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, ", "))
	}
	return nil
}
func (r *Rule) validateUsers() (string, error) {
	msgs := []string{}
	for _, u := range r.FileList.Users() {
		if _, err := user.Lookup(u); err != nil {
			if _, ok := err.(user.UnknownUserError); ok {
				msgs = append(msgs, fmt.Sprintf("user `%s` not found", u))
			} else {
				return "", fmt.Errorf("can not lookup user: %s: %s", u, err)
			}
		}
	}
	return strings.Join(msgs, ", "), nil
}
func (r *Rule) validateGroups() (string, error) {
	// TODO: waiting for https://github.com/golang/go/issues/2617
	return "", nil
}
func (r *Rule) validateROOT() (string, error) {
	msgs := []string{}
	lists := map[string]struct{}{}
	files := map[string]struct{}{}
	dirs := map[string]struct{}{}

	for _, fi := range r.FileList {
		lists[fi.Path] = struct{}{}
	}
	err := filepath.Walk(r.rootPath(), func(path string, info os.FileInfo, err error) error {
		p := strings.TrimPrefix(path, r.rootPath())
		if info.IsDir() {
			dirs[p] = struct{}{}
		}
		if p != "" {
			files[p] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to scan files: %s", err)
	}

	for _, fi := range r.FileList {
		for dir := range dirs {
			if fi.Path == dir {
				fi.IsDir = true
			}
		}
	}
	for f := range lists {
		if _, ok := files[f]; !ok {
			msgs = append(msgs, fmt.Sprintf("ROOT dir not included: %s", f))
		}
	}
	return strings.Join(msgs, ", "), nil
}

func (r *Rule) Install(dryrun bool) error {
	for _, name := range r.Hooks.AtFirst {
		s := r.Hooks.Scripts[name]
		logrus.Infof("HOOK(at_first): [%s] %s", name, s.Cmd)
		if err := s.Run(dryrun); err != nil {
			return fmt.Errorf("failed to execute `%s` script at first hook: %s", name, err)
		}
	}

	onceHooks := map[string]*script{}
	for _, fi := range r.FileList {
		changed, err := r.install(fi, dryrun)
		if err != nil {
			return fmt.Errorf("failed to install %s: %s", fi.SrcPath, err)
		}

		if changed {
			for _, name := range r.Hooks.Changed[fi.Path] {
				s := r.Hooks.Scripts[name]
				if s.Once {
					onceHooks[name] = s
				} else {
					logrus.Infof("HOOK(changed): [%s] %s", name, s.Cmd)
					if err := s.Run(dryrun); err != nil {
						return fmt.Errorf("failed to execute `%s` script at changed hook: %s", name, err)
					}
				}
			}
		}
	}
	for n, s := range onceHooks {
		logrus.Infof("HOOK(changed): [%s] %s", n, s.Cmd)
		if err := s.Run(dryrun); err != nil {
			return fmt.Errorf("failed to execute `%s` script at changed hook: %s", n, err)
		}
	}
	for _, name := range r.Hooks.AtLast {
		s := r.Hooks.Scripts[name]
		logrus.Infof("HOOK(at_last): [%s] %s", name, s.Cmd)
		if err := s.Run(dryrun); err != nil {
			return fmt.Errorf("failed to execute `%s` script at last hook: %s", name, err)
		}
	}
	return nil
}

func (r *Rule) install(fi *FileInfo, dryrun bool) (bool, error) {
	changed := false
	opt := install.InstallOption{
		Owner: fi.Owner,
		Group: fi.Group,
		Mode:  &fi.Mode,
	}
	if fi.IsDir {
		differ, err := HasDirDiff(fi.Path, fi.Owner, fi.Group, fi.Mode)
		if err != nil {
			return changed, err
		}
		if differ {
			fmt.Println(fi.String())
			changed = true
		}
		if !dryrun && differ {
			if err := install.InstallDir(fi.Path, &opt); err != nil {
				return changed, err
			}
		}
	} else {
		differ, err := HasFileDiff(fi.SrcPath, fi.Path)
		if err != nil {
			return changed, err
		}
		if differ {
			fmt.Println(fi.String())
			changed = true
		}
		if !dryrun && differ {
			if err := install.InstallFile(fi.SrcPath, fi.Path, &opt); err != nil {
				return changed, err
			}
		}
	}
	return changed, nil
}
