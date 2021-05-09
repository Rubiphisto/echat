package tcp

import (
	"context"
	"net"
	"sync"
	"time"
	
	"echat/utils/logger"
)

// Server Tcp服务器对象
type Server interface {
	// Start 启动Tcp服务器
	Start(context context.Context, group *sync.WaitGroup) error
	// Stop 停止Tcp服务器
	Stop()
	// GetHeartbeatInterval 获取连接心跳检测间隔时间
	GetHeartbeatInterval() time.Duration
}

type tcpServer struct {
	listener          net.Listener
	factory           SessionFactory
	serialFactory     SerializeFactory
	maxConnectionId   uint32
	mutex             sync.Mutex
	connections       map[uint32]*connection
	connectionGroup   sync.WaitGroup
	heartbeatInterval time.Duration
	context           context.Context
	contextCancel     context.CancelFunc
}

// NewTcpServer 构建Tcp服务器对象
func NewTcpServer(addr string, factory SessionFactory, serialFactory SerializeFactory, heartbeatInterval time.Duration) (Server, error) {
	listener, err := net.Listen("tcp", addr)
	if nil != err {
		return nil, err
	}

	logger.Info("Server listen on %v", addr)

	return &tcpServer{
		listener:          listener,
		factory:           factory,
		serialFactory:     serialFactory,
		maxConnectionId:   0,
		connections:       make(map[uint32]*connection),
		heartbeatInterval: heartbeatInterval,
	}, nil
}

func (s *tcpServer) Start(ctx context.Context, group *sync.WaitGroup) error {
	group.Add(1)
	s.context, s.contextCancel = context.WithCancel(ctx)
	go s.run(group)
	return nil
}

func (s *tcpServer) run(group *sync.WaitGroup) {
	logger.Info("Start the tcp server accept routine")
	defer func() {
		// 关闭所有网络连接，并等待完成
		for _, conn := range s.connections {
			conn.Stop()
		}
		s.connectionGroup.Wait()
		group.Done()
	}()

	errCount := 0

	for {
		conn, err := s.listener.Accept()
		if nil != err {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				errCount++
				index := errCount
				if index >= len(acceptRetryInterval) {
					index = len(acceptRetryInterval) - 1
				}
				time.Sleep(acceptRetryInterval[index])
				continue
			}
			select {
			case <-s.context.Done():
				logger.Info("Server Accept quit with done")
				return
			default:
				logger.Error("Server Accept quit with error: %v", err)
				return
			}
		}
		errCount = 0

		connection, err := NewConnection(s.context,
			conn,
			s.factory.CreateSession(),	
			s.serialFactory.CreateSerializer(),	
			s.serialFactory.CreateDeserializer(),	
			s.heartbeatInterval)
		if nil == connection || nil != err {
			conn.Close()
			logger.Error("Failed to construct connection for %v", conn.RemoteAddr())
			continue
		}

		s.addConnection(connection)

		s.connectionGroup.Add(1)
		go func() {
			defer func() {
				s.delConnection(connection.connectionId)
				s.connectionGroup.Done()
			}()

			connection.run()
			logger.Debug("Server the routine of connection %v has exited.", connection.connectionId)
		}()
	}
}

func (s *tcpServer) GetHeartbeatInterval() time.Duration {
	return s.heartbeatInterval
}

func (s *tcpServer) addConnection(connection *connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.maxConnectionId += 1

	connection.connectionId = s.maxConnectionId
	s.connections[connection.connectionId] = connection
}

func (s *tcpServer) delConnection(connectionId uint32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	connection, ok := s.connections[connectionId]
	if !ok {
		return
	}

	connection.Stop()
	delete(s.connections, connectionId)
}

func (s *tcpServer) Stop() {
	s.contextCancel()
	s.listener.Close()
}
