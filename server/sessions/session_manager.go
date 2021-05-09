package sessions

import (
	"context"
	"echat/utils/tcp"
	"encoding/binary"
	"sync"
	"time"
)

type SessionManager struct {
	tcpServer		tcp.Server
}

var (
	sessionManager SessionManager
)

func init() {
}

func GetSessionManager() *SessionManager {
	return &sessionManager
}

func (m *SessionManager) Start(ctx context.Context, wg *sync.WaitGroup) error {
	server, err := tcp.NewTcpServer("0.0.0.0:10002", m, tcp.GetDefaultSerializeFactory(binary.LittleEndian), time.Second * 5)
	if nil != err {
		return err
	}
	m.tcpServer = server
	return m.tcpServer.Start(ctx, wg)
}

func (m *SessionManager) Stop() {
	m.tcpServer.Stop()
}

func (m *SessionManager) CreateSession() tcp.Session {
	return NewSession()
}

