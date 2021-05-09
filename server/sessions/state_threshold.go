package sessions

import (
	"echat/common/pb"
	"google.golang.org/protobuf/proto"
)

type SessionStateThreshold struct {
	SessionState
}

func NewStateThreshold(name string, session *Session) State {
	state := &SessionStateThreshold{}
	state.Initialize(session, name)
	return state
}

func (s *SessionStateThreshold) OnEnter() {
	_ = s.AddHandler(pb.MessageId_LoginRequest, s.onLoginRequest)
}

func (s *SessionStateThreshold) OnExit() {
	s.DelHandler(pb.MessageId_LoginRequest)
}

func (s *SessionStateThreshold) onLoginRequest(_ uint32, data []byte) error {
	req := &pb.LoginRequestMessage{}
	if err := proto.Unmarshal(data, req); nil != err {
		return err
	}

	resp := &pb.LoginResponseMessage{}
	
	user := GetUserManager().GetUser(req.Username)
	if nil != user {
		resp.Result = pb.Result_DuplicatedName
	} else {
		user = GetUserManager().CreateUser(req.Username, s.GetSession())
		if nil != user {
			s.GetSession().username = user.GetUserName()
			s.GetSession().Translate("Lobby")
			resp.Result = pb.Result_Success
		} else {
			resp.Result = pb.Result_Error
		}
	}
	s.SendMessage(pb.MessageId_LoginResponse, resp)

	return nil
}

