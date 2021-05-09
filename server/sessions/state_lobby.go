package sessions

import (
	"echat/common/pb"
	"echat/utils/logger"
	"google.golang.org/protobuf/proto"
)

type SessionStateLobby struct {
	SessionState
}

func NewStateLobby(name string, session *Session) State {
	state := &SessionStateLobby{}
	state.Initialize(session, name)
	return state
}

func (s *SessionStateLobby) OnEnter() {
	_ = s.AddHandler(pb.MessageId_EnterChannelRequest, s.onEnterChannel)
	logger.Info("user %v enter lobby", s.GetSession().username)
}

func (s *SessionStateLobby) OnExit() {
	s.DelHandler(pb.MessageId_EnterChannelRequest)
}

func (s *SessionStateLobby) onEnterChannel(_ uint32, data []byte) error {
	req := &pb.EnterChannelRequestMessage{}
	if err := proto.Unmarshal(data, req); nil != err {
		return err
	}
	
	if 0 == len(s.GetSession().username) {
		s.GetSession().Translate("Threshold")
		return nil
	}
	
	user := GetUserManager().GetUser(s.GetSession().username)
	if nil == user {
		s.SendMessage(pb.MessageId_EnterChannelResponse, &pb.EnterChannelResponseMessage{
			Result:		pb.Result_NotFoundUser,
		})
		return nil
	}
	if user.IsInChannel() {
		s.SendMessage(pb.MessageId_EnterChannelResponse, &pb.EnterChannelResponseMessage{
			Result:		pb.Result_AlreadyInChannel,
		})
		return nil
	}
	
	GetChannelManager().EnterChannel(user, req.ChannelName)
	s.GetSession().Translate("Channel")
	return nil
}
