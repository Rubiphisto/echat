package session

import (
	"context"
	"echat/common/pack"
	"echat/utils/logger"
	"echat/utils/tcp"
	"encoding/binary"
	"fmt"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)
// MessageHandler 游戏服消息处理器
type MessageHandler func(msgId uint32, data []byte) error

type Session struct {
	tcpClient	tcp.Client
	id   		uint32
	connection 	tcp.Connection
	
	handlers 	map[uint32]MessageHandler
	state		State
	username	string
	channelName	string
}

func NewSession() *Session {
	return &Session{
		handlers: map[uint32]MessageHandler{},
	}
}

func (m *Session) Start(ctx context.Context, wg *sync.WaitGroup) error {
	client, err := tcp.NewTcpClient("127.0.0.1:10002", m, tcp.GetDefaultSerializeFactory(binary.LittleEndian), time.Second * 5)
	if nil != err {
		return err
	}
	m.tcpClient = client
	return m.tcpClient.Start(ctx, wg)
}

func (m *Session) Stop() {
	m.tcpClient.Stop()
}

func (m *Session) CreateSession() tcp.Session {
	return m
}

// Initialize 连接建立后被调用
func (m *Session) Initialize(connection tcp.Connection) error {
	m.id = connection.GetConnectionId()
	m.connection = connection
	logger.Info("session.%v Initialize", m.id)
	if err := m.Translate("Threshold"); nil != err {
		return err
	}
	return nil
}

// Uninitialized 连接关闭后被调用
func (m *Session) Uninitialized() {
	logger.Info("session.%v Uninitialized", m.id)
}

// OnRecvMessage 收到数据包
func (m *Session) OnRecvMessage(content []byte) {
	pack, err := pack.Unpack(content)
	if nil != err {
		logger.Error("Parse msg pack error, terminate the connection")
		m.connection.Stop()
		return
	}

	handler, ok := m.handlers[pack.MsgId]
	if !ok {
		logger.Info("Tcp sesssion drop unhandled msgID %d", pack.MsgId)
		return
	}
	if err := handler(pack.MsgId, pack.Data); nil != err {
		logger.Error("Failed to handle message %v with error %v", pack.MsgId, err.Error())
	}
}

// CheckHeartbeat 心跳检测，返回 false 表示断开网络连接
func (m *Session) CheckHeartbeat() bool {
	return true
}

// Translate 切换会话状态
func (m *Session) Translate(name string) error {
	if nil != m.state && m.state.GetName() == name {
		return nil
	}

	nextState := myFactory.CreateState(name, m)
	if nil == nextState {
		return fmt.Errorf("failed to create state '%v'", name)
	}
	nextState.OnCreate()
	prevState := m.state
	m.state = nextState
	if nil != prevState {
		prevState.OnExit()
	}
	m.state.OnEnter()

	return nil
}

// GetConnection 获得会话绑定网络连接对象
func (m *Session) GetConnection() tcp.Connection {
	return m.connection
}

// SendMessage 发送消息
func (m *Session) SendMessage(msgId uint32, msg proto.Message) bool {
	data, err := proto.Marshal(msg)
	if nil != err {
		return false
	}
	p := &pack.MsgPack{
		Length:  0,
		MsgId:   msgId,
		Data: 	 data,
	}
	b, err := pack.Pack(p)
	if nil != err {
		return false
	}
	return m.GetConnection().Send(b)
}

// AddHandler 注册网络消息处理器
func (m *Session) AddHandler(msgId uint32, handler MessageHandler) error {
	if _, ok := m.handlers[msgId]; ok {
		return fmt.Errorf("the handler of msgId %v is exist", msgId)
	}
	m.handlers[msgId] = handler
	return nil
}

// DelHandler 移除网络消息处理器
func (m *Session) DelHandler(msgId uint32) {
	delete(m.handlers, msgId)
}



