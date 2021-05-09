package tcp

import (
	"context"
	"echat/utils/logger"
	"net"
	"sync"
	"time"
)

type Client interface {
	// Start 启动Tcp客户端
	Start(context context.Context, group *sync.WaitGroup) error
	// Stop 停止Tcp客户端
	Stop()
	// GetHeartbeatInterval 获取连接心跳检测间隔时间
	GetHeartbeatInterval() time.Duration
}

type tcpClient struct {
	addr              string
	connection        Connection
	factory           SessionFactory
	serialFactory     SerializeFactory
	heartbeatInterval time.Duration
	context           context.Context
	contextCancel     context.CancelFunc
}

// NewTcpClient 构建Tcp客户端连接对象
func NewTcpClient(addr string, factory SessionFactory, serialFactory SerializeFactory, heartbeatInterval time.Duration) (Client, error) {
	return &tcpClient{
		addr:					addr,
		factory:				factory,
		serialFactory:			serialFactory,
		heartbeatInterval:		heartbeatInterval,
	}, nil
}

func (c *tcpClient) Start(ctx context.Context, waitGroup *sync.WaitGroup) error {
	c.context, c.contextCancel = context.WithCancel(ctx)
	return c.run(waitGroup)
}

func (c *tcpClient) Stop() {
	c.contextCancel()
}

func (c *tcpClient) GetHeartbeatInterval() time.Duration {
	return c.heartbeatInterval
}

func (c *tcpClient) run(waitGroup *sync.WaitGroup) error {
	conn, err := net.Dial("tcp", c.addr)
	if nil != err {
		return err
	}
	connection, err := NewConnection(c.context,
		conn,
		c.factory.CreateSession(),
		c.serialFactory.CreateSerializer(),
		c.serialFactory.CreateDeserializer(),
		c.heartbeatInterval)
	if nil == connection || nil != err {
		conn.Close()
		logger.Error("Failed to construct connection for %v", conn.RemoteAddr())
		return err
	}
	connection.connectionId = 1
	c.connection = connection
	
	waitGroup.Add(1)
	go func() {
		defer func() {
			c.connection = nil
			waitGroup.Done()
		}()
		
		connection.run()
		logger.Debug("Client the routine of connection %v has exited.", connection.connectionId)
	}()
	
	return nil
}
