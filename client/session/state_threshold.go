package session

import (
	"echat/client/console"
	"echat/common/pb"
	"echat/utils/logger"
	"fmt"
	"google.golang.org/protobuf/proto"
)

type SessionStateThreshold struct {
	SessionState
	
	consoleId		uint32
}

func NewStateThreshold(name string, session *Session) State {
	state := &SessionStateThreshold{}
	state.Initialize(session, name)
	return state
}

func (s *SessionStateThreshold) OnEnter() {
	_ = s.AddHandler(pb.MessageId_LoginResponse, s.onLoginResponse)
	console.NewConsole().AddHandler("login", s.cmdLogin)
}

func (s *SessionStateThreshold) OnExit() {
	s.DelHandler(pb.MessageId_LoginResponse)
	console.NewConsole().DelHandler("login")
}

func (s *SessionStateThreshold) cmdLogin(params []string) {
	if 0 == len(params) {
		logger.Error("no username")
		return
	}
	req := &pb.LoginRequestMessage{
		Username: params[0],
	}
	s.SendMessage(pb.MessageId_LoginRequest, req)
}

func (s *SessionStateThreshold) onLoginResponse(_ uint32, data []byte) error {
	resp := &pb.LoginResponseMessage{}
	if err := proto.Unmarshal(data, resp); nil != err {
		return err
	}
	fmt.Printf("login to server with result %v\n", resp.Result)
	if pb.Result_Success == resp.Result {
		s.GetSession().Translate("Lobby")
	}
	return nil
}

