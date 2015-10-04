package barrow

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"syscall"

	"github.com/naegelejd/go-acl/os/group"
)

func HasDirDiff(path, username, groupname string, mode os.FileMode) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, errors.New("not a directory")
	}
	if (fi.Mode() - os.ModeDir) != mode {
		return true, nil
	}

	uid := fi.Sys().(*syscall.Stat_t).Uid
	gid := fi.Sys().(*syscall.Stat_t).Gid
	u, err := user.LookupId(fmt.Sprint(uid))
	if err != nil {
		return false, err
	}
	g, err := group.LookupId(fmt.Sprint(gid))
	if err != nil {
		return false, err
	}
	if u.Username != username {
		return true, nil
	}
	if g.Name != groupname {
		return true, nil
	}
	return false, nil
}

func HasFileDiff(aPath, bPath string) (bool, error) {
	a, err := sha256File(aPath)
	if err != nil {
		return false, err
	}
	b, err := sha256File(bPath)
	if err != nil {
		return false, err
	}
	return a != b, nil
}

func sha256File(path string) (string, error) {
	h := sha256.New()
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%b", h.Sum(nil)), nil
}
