package agent

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var defaultNameservers = []string{
	"8.8.8.8",
	"8.8.4.4",
}

func NewResolvConf(ns []string) *ResolvConf {
	if len(ns) == 0 {
		ns = defaultNameservers
	}
	return &ResolvConf{
		Nameservers: ns,
	}
}

type ResolvConf struct {
	Nameservers []string
}

func (r *ResolvConf) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0711); err != nil {
		return err
	}
	f, err := ioutil.TempFile("", "orbit-resolvconf")
	if err != nil {
		return err
	}
	if err := f.Chmod(0666); err != nil {
		return err
	}
	for _, ns := range r.Nameservers {
		if _, err := f.WriteString(fmt.Sprintf("nameserver %s\n", ns)); err != nil {
			f.Close()
			return err
		}
	}
	f.Close()
	return os.Rename(f.Name(), path)
}

func setupHostResolvConf() error {
	r := NewResolvConf([]string{"127.0.0.1"})
	return r.Write("/etc/resolv.conf")
}
