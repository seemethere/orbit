package agent

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	lconfig "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/server"
	v1 "github.com/stellarproject/orbit/api/v1"
)

func newStore(root string) (*server.App, error) {
	cfg := lconfig.NewConfigDefault()
	cfg.Addr = fmt.Sprintf("0.0.0.0:%d", 6379)
	cfg.UseReplication = true
	cfg.DataDir = filepath.Join(root, "ledis")
	server, err := server.NewApp(cfg)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		server.Run()
	}()
	wg.Wait()
	return server, nil
}

func newStoreClient() *store {
	return &store{
		pool: redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		}, 5),
	}
}

const (
	serviceFormat = "orbit:%s"
)

type store struct {
	pool *redis.Pool
}

func (s *store) Close() error {
	return s.pool.Close()
}

func (s *store) Register(id string, service *v1.Service) error {
	data, err := proto.Marshal(service)
	if err != nil {
		return err
	}
	_, err = s.do("HSET", fmt.Sprintf(serviceFormat, service.Name), id, data)
	return err
}

func (s *store) Deregister(id, name string) error {
	name = fmt.Sprintf(serviceFormat, name)
	if _, err := s.do("HDEL", name, id); err != nil {
		return err
	}
	l, err := redis.Int(s.do("HLEN", name))
	if err != nil {
		return err
	}
	if l == 0 {
		_, err = s.do("DEL", name)
		return err
	}
	return nil
}

func (s *store) do(cmd string, args ...interface{}) (interface{}, error) {
	conn := s.pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}
