package agent

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type key struct {
	Domain string
	Name   string
	ID     string
}

func (k key) isEmpty() bool {
	return k.Domain == ""
}

func getKey(q dns.Question) key {
	parts := strings.Split(q.Name, ".")
	switch len(parts) {
	case 3:
		return key{
			Domain: parts[1],
			Name:   parts[0],
		}
	case 4:
		return key{
			Domain: parts[2],
			Name:   parts[1],
			ID:     parts[0],
		}
	}
	return key{}
}

func createHeader(rtype uint16, name string, ttl uint32) dns.RR_Header {
	return dns.RR_Header{
		Name:   name,
		Rrtype: rtype,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
}

func getProto(w dns.ResponseWriter) string {
	proto := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		proto = "tcp"
	}
	return proto
}

func startDnsServer(mux *dns.ServeMux, net, addr string, udpsize int, timeout time.Duration) {
	s := &dns.Server{
		Addr:         addr,
		Net:          net,
		Handler:      mux,
		UDPSize:      udpsize,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	if err := s.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("listen")
	}
}
