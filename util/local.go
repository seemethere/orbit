package util

import (
	v1 "github.com/stellarproject/orbit/api/v1"
	"google.golang.org/grpc"
)

type LocalAgent struct {
	v1.AgentClient
	v1.DHCPServer
	conn *grpc.ClientConn
}

func (a *LocalAgent) Close() error {
	return a.conn.Close()
}

func Agent(address string) (*LocalAgent, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &LocalAgent{
		AgentClient: v1.NewAgentClient(conn),
		conn:        conn,
	}, nil
}
