package barrow

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type FileInfo struct {
	Mode    os.FileMode
	Owner   string
	Group   string
	SrcPath string
	Path    string
	IsDir   bool
}

func (fi *FileInfo) String() string {
	return fmt.Sprintf("%s %s %s %s", fmt.Sprintf("%o", fi.Mode), fi.Owner, fi.Group, fi.Path)
}

type FileList []*FileInfo

func ParseFileList(path string) (list FileList, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	lines := bytes.Split(b, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		v := bytes.SplitN(line, []byte(" "), 4)
		m, err := strconv.ParseInt(string(v[0]), 8, 32)
		if err != nil {
			return nil, err
		}
		list = append(list, &FileInfo{
			Mode:  os.FileMode(m),
			Owner: string(v[1]),
			Group: string(v[2]),
			Path:  string(v[3]),
		})
	}
	return
}

func (f FileList) Users() []string {
	users := []string{}
	m := map[string]struct{}{}
	for _, fi := range f {
		if _, ok := m[fi.Owner]; !ok {
			users = append(users, fi.Owner)
		} else {
			m[fi.Owner] = struct{}{}
		}
	}
	return users
}

func (f FileList) Groups() []string {
	users := []string{}
	m := map[string]struct{}{}
	for _, fi := range f {
		if _, ok := m[fi.Owner]; !ok {
			users = append(users, fi.Group)
		} else {
			m[fi.Group] = struct{}{}
		}
	}
	return users
}
