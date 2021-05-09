package sessions

import (
	"echat/common/pb"
	"echat/utils/logger"
	"google.golang.org/protobuf/proto"
)

type SessionStateChannel struct {
	SessionState
}
func NewStateChannel(name string, session *Session) State {
	state := &SessionStateChannel{}
	state.Initialize(session, name)
	return state
}

func (s *SessionStateChannel) OnEnter() {
	_ = s.AddHandler(pb.MessageId_ChatRequest, s.onChat)
	_ = s.AddHandler(pb.MessageId_LeaveChannelRequest, s.onLeaveChannel)
	logger.Info("user %v enter channel", s.GetSession().username)
}

func (s *SessionStateChannel) OnExit() {
	s.DelHandler(pb.MessageId_ChatRequest)
	s.DelHandler(pb.MessageId_LeaveChannelRequest)

	user := GetUserManager().GetUser(s.GetSession().username)
	if nil == user {
		return
	}
	channel := GetChannelManager().GetChannel(user.GetChannelName())
	if nil == channel {
		return
	}
	channel.DelUser(user)
}

func (s *SessionStateChannel) onChat(_ uint32, data []byte) error {
	req := &pb.ChatRequestMessage{}
	if err := proto.Unmarshal(data, req); nil != err {
		return err
	}
	user := GetUserManager().GetUser(s.GetSession().username)
	if nil == user {
		return nil
	}
	channel := GetChannelManager().GetChannel(user.GetChannelName())
	if nil == channel {
		s.GetSession().Translate("Lobby")
		return nil
	}
	channel.Chat(user.GetUserName(), req.Message)
	return nil
}

func (s *SessionStateChannel) onLeaveChannel(_ uint32, data []byte) error {
	req := &pb.LeaveChannelRequestMessage{}
	if err := proto.Unmarshal(data, req); nil != err {
		return err
	}
	s.GetSession().Translate("Lobby")
	return nil
}

