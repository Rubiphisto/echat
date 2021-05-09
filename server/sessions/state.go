package sessions

import (
	"fmt"

	"echat/common/pb"
	"echat/utils/tcp"
	
	"google.golang.org/protobuf/proto"
)

// State 会话状态基础类
type State interface {
	// GetName 获取状态名字
	GetName() string
	// OnCreate 状态创建时调用
	OnCreate()
	// OnEnter 状态进入
	OnEnter()
	OnExit()
	OnDestroy()
}

type StateCreator func(name string, session *Session) State

type StateFactory struct {
	stateCreators map[string]StateCreator
}

func (s *StateFactory) RegisterCreator(name string, creator StateCreator) {
	if _, ok := s.stateCreators[name]; ok {
		panic(fmt.Sprintf("state %v is already exist", name))
	}
	s.stateCreators[name] = creator
}

func (s *StateFactory) CreateState(name string, session *Session) State {
	creator, ok := s.stateCreators[name]
	if !ok {
		return nil
	}
	return creator(name, session)
}

var myFactory *StateFactory

func init() {
	myFactory = &StateFactory{stateCreators: map[string]StateCreator{}}
	myFactory.RegisterCreator("Threshold", NewStateThreshold)
	myFactory.RegisterCreator("Lobby", NewStateLobby)
	myFactory.RegisterCreator("Channel", NewStateChannel)
}

// region: SessionState
type SessionState struct {
	name    string
	session *Session
}

func (s *SessionState) Initialize(session *Session, name string) {
	s.name = name
	s.session = session
}

func (s *SessionState) OnCreate() {
}

func (s *SessionState) GetName() string {
	return s.name
}

func (s *SessionState) OnEnter() {
}

func (s *SessionState) OnExit() {
}

func (s *SessionState) OnDestroy() {
}

func (s *SessionState) GetSession() *Session {
	return s.session
}

func (s *SessionState) GetConnection() tcp.Connection {
	return s.session.GetConnection()
}

func (s *SessionState) AddHandler(msgId pb.MessageId, handler MessageHandler) error {
	return s.session.AddHandler(uint32(msgId), handler)
}

func (s *SessionState) DelHandler(msgId pb.MessageId) {
	s.session.DelHandler(uint32(msgId))
}

func (s *SessionState) SendMessage(msgId pb.MessageId, msg proto.Message) bool {
	return s.session.SendMessage(uint32(msgId), msg)
}

// endregion: SessionState


