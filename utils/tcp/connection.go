package tcp

import (
	"context"
	"net"
	"sync"
	"time"

	"echat/utils/logger"
	utilTime "echat/utils/time"
)

const (
	// sendSizeLimit send size limit, close connection when pending overflow the limit
	sendSizeLimit = 10240
	// readSizeLimit send size limit, close connection when pending overflow the limit
	readSizeLimit = 10240
	// Heartbeat check interval
	minHeartCheckInterval = time.Second * 1
)

var acceptRetryInterval = []time.Duration{
	time.Second,
	time.Second * 2,
	time.Second * 5,
	time.Second * 15,
	time.Second * 30,
	time.Second * 60,
}

// Connection 网络连接对象
type Connection interface {
	// GetConnectionId 获取网络连接号
	GetConnectionId() uint32
	// Send 发送数据包
	Send(data []byte) bool
	// Stop 关停网络连接
	Stop()
	// ScheduleTask 注册计划回调任务
	ScheduleTask(duration time.Duration, isTicker bool, callback utilTime.SchedulerCallback) (scheduleId uint64, err error)
	// UnscheduleTask 取消指定计划回调
	UnscheduleTask(scheduleId uint64) error
}

type connection struct {
	conn          net.Conn
	server        Server
	session       Session
	connectionId  uint32
	context       context.Context
	contextCancel context.CancelFunc

	wait   sync.WaitGroup
	sender chan []byte
	reader chan []byte

	scheduler    utilTime.Scheduler
	serializer   ConnectSerializer
	deserializer ConnectDeserializer
	heartbeat    time.Duration
}

func NewConnection(ctx context.Context, conn net.Conn, session Session, serial ConnectSerializer, deserial ConnectDeserializer, heartbeat time.Duration) (*connection, error) {
	connection := &connection{
		conn:         conn,
		session:      session,
		sender:       make(chan []byte, sendSizeLimit),
		reader:       make(chan []byte, readSizeLimit),
		scheduler:    utilTime.NewScheduler(),
		serializer:   serial,
		deserializer: deserial,
		heartbeat:    heartbeat,
	}
	connection.context, connection.contextCancel = context.WithCancel(ctx)
	if err := connection.scheduler.Start(connection.context, &connection.wait); nil != err {
		return nil, err
	}
	return connection, nil
}

func (c *connection) GetConnectionId() uint32 {
	return c.connectionId
}

func (c *connection) run() {
	logger.Info("Tcp connection run, remoteAddr: %s", c.conn.RemoteAddr())

	if err := c.session.Initialize(c); nil != err {
		logger.Error("Failed to initialize the session on connection, %v", c.conn.RemoteAddr())
		close(c.sender)
		close(c.reader)
		_ = c.conn.Close()
		return
	}

	// start the send/recv routine
	c.wait.Add(1)
	go c.sendRoutine()
	c.wait.Add(1)
	go c.recvRoutine()

	interval := c.heartbeat
	if interval < minHeartCheckInterval {
		interval = minHeartCheckInterval
	}
	ticker := time.NewTicker(interval)
	// execute the connection loop
	func() {
		for {
			select {
			case <-c.context.Done():
				return
			case content := <-c.reader:
				c.session.OnRecvMessage(content)
			case deliver := <-c.scheduler.Done():
				deliver.Call()
			case <-ticker.C:
				if !c.session.CheckHeartbeat() {
					logger.Info("the heartbeat check of connection %v is failed, shutdown the connection", c.GetConnectionId())
					c.Stop()
				}
			}
		}
	}()

	ticker.Stop()

	// wait for send/recv routine to terminate
	_ = c.conn.SetDeadline(time.Now())
	close(c.sender)
	close(c.reader)
	c.scheduler.Stop()
	c.wait.Wait()

	// cleanup
	_ = c.conn.Close()
	c.session.Uninitialized()
}

func (c *connection) Send(data []byte) bool {
	select {
	case <-c.context.Done():
		return false
	case c.sender <- data:
		return true
	default:
		logger.Info("Sender conn: %s, sender %d/%d, close connection\n", c.conn.RemoteAddr(), len(c.sender), cap(c.sender))
		c.Stop()
		return false
	}
}

func (c *connection) sendRoutine() {
	defer c.wait.Done()

	for data := range c.sender {
		// 连接停止时会先退出 main routine 循环，主循环中关闭 sender，该循环退出
		select {
		case <-c.context.Done():
			logger.Info("stop send routine with Done")
			return
		default:
			if err := c.rawSend(data); nil != err {
				logger.Info("rawSend error: %v", err)
				c.Stop()
				return
			}
		}
	}
}

func (c *connection) rawSend(data []byte) error {
	return c.serializer.Serialize(c.connectionId, c.conn, data)
}

func (c *connection) recvRoutine() {
	defer c.wait.Done()
	for {
		if err := c.conn.SetReadDeadline(time.Time{}); nil != err {
			logger.Info("Tcp connection SetReadDeadline error: %v", err)
		}

		content, err := c.deserializer.Deserialize(c.connectionId, c.conn)
		if nil != err {
			logger.Info("TcpConnection|read error: %v", err)
			c.Stop()
			return
		}
		c.reader <- content
	}
}

func (c *connection) Stop() {
	c.contextCancel()
}

// ScheduleTask 注册计划回调任务
func (c *connection) ScheduleTask(duration time.Duration, isTicker bool, callback utilTime.SchedulerCallback) (scheduleId uint64, err error) {
	return c.scheduler.Schedule(duration, isTicker, callback)
}

// UnscheduleTask 取消指定计划回调
func (c *connection) UnscheduleTask(scheduleId uint64) error {
	return c.scheduler.Unschedule(scheduleId)
}

