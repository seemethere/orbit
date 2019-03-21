package util

import (
	"errors"
	"net"
	"os"
	"runtime"
	"syscall"
)

var ErrIPAddressNotFound = errors.New("box: ip address for interface not found")

func GetIP(name string) (string, error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	return getIPf(i, ipv4)
}

func getIPf(i *net.Interface, ipfunc func(n *net.IPNet) string) (string, error) {
	addrs, err := i.Addrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		n, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		s := ipfunc(n)
		if s == "" {
			continue
		}
		return s, nil
	}
	return "", ErrIPAddressNotFound
}

func ipv4(n *net.IPNet) string {
	if n.IP.To4() == nil {
		return ""
	}
	return n.IP.To4().String()
}

func GetDomainName() (name string, err error) {
	// Try uname first, as it's only one system call and reading
	// from /proc is not allowed on Android.
	var un syscall.Utsname
	err = syscall.Uname(&un)

	var buf [512]byte // Enough for a DNS name.
	for i, b := range un.Domainname[:] {
		buf[i] = uint8(b)
		if b == 0 {
			name = string(buf[:i])
			break
		}
	}
	// If we got a name and it's not potentially truncated
	// (Nodename is 65 bytes), return it.
	if err == nil && len(name) > 0 && len(name) < 64 {
		return name, nil
	}
	if runtime.GOOS == "android" {
		if name != "" {
			return name, nil
		}
		return "localdomain", nil
	}

	f, err := os.Open("/proc/sys/kernel/domainname")
	if err != nil {
		return "", err
	}
	defer f.Close()

	n, err := f.Read(buf[:])
	if err != nil {
		return "", err
	}

	if n > 0 && buf[n-1] == '\n' {
		n--
	}
	return string(buf[:n]), nil
}
