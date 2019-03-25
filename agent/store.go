package agent

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	lconfig "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/server"
	"github.com/sirupsen/logrus"
	v1 "github.com/stellarproject/orbit/api/v1"
	"github.com/stellarproject/orbit/util"
)

const (
	serviceFormat  = "orbit:%s"
	storePort      = 6379
	localStoreAddr = "127.0.0.1:6379"
	masterDomain   = "master"
)

func newStore(root string, master bool) (*server.App, error) {
	cfg := lconfig.NewConfigDefault()
	cfg.Addr = fmt.Sprintf("0.0.0.0:%d", storePort)
	cfg.UseReplication = true
	cfg.DataDir = filepath.Join(root, "ledis")
	cfg.Readonly = !master

	server, err := server.NewApp(cfg)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		logrus.WithField("address", cfg.Addr).Infof("starting store")
		wg.Done()
		server.Run()
	}()
	wg.Wait()
	return server, nil
}

func newStoreClient(master string) *store {
	if master == "" {
		master = localStoreAddr
	}
	m := redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", master)
	}, 5)
	var slave *redis.Pool
	if master == "" {
		slave = m
	} else {
		slave = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", localStoreAddr)
		}, 5)
	}
	return &store{
		m: m,
		s: slave,
	}
}

type store struct {
	m *redis.Pool
	s *redis.Pool
}

func (s *store) Close() error {
	s.s.Close()
	return s.m.Close()
}

func (s *store) RegisterMaster(iface string) error {
	ip, err := util.GetIP(iface)
	if err != nil {
		return err
	}
	service := &v1.Service{
		Name: masterDomain,
		IP:   ip,
		Port: storePort,
	}
	return s.Register("store", service)
}

func (s *store) Register(id string, service *v1.Service) error {
	data, err := proto.Marshal(service)
	if err != nil {
		return err
	}
	_, err = s.master("HSET", fmt.Sprintf(serviceFormat, service.Name), id, data)
	return err
}

func (s *store) Deregister(id, name string) error {
	name = fmt.Sprintf(serviceFormat, name)
	if _, err := s.master("HDEL", name, id); err != nil {
		return err
	}
	l, err := redis.Int(s.master("HLEN", name))
	if err != nil {
		return err
	}
	if l == 0 {
		_, err = s.master("DEL", name)
		return err
	}
	return nil
}

func (s *store) fetchService(name string) (*v1.Service, error) {
	values, err := redis.Values(s.slave("HGETALL", fmt.Sprintf(serviceFormat, name)))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.Errorf("service %q does not exist", name)
	}
	var v []byte
	for _, val := range values {
		if val == nil {
			continue
		}
		v = val.([]byte)
	}
	if v == nil {
		return nil, errors.Errorf("service %q does not exist", name)
	}
	var service v1.Service
	if err := proto.Unmarshal(v, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

func (s *store) slave(cmd string, args ...interface{}) (interface{}, error) {
	conn := s.s.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (s *store) master(cmd string, args ...interface{}) (interface{}, error) {
	conn := s.m.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}
