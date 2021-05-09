package session

import (
	"echat/client/console"
	"echat/common/pb"
	"echat/utils/logger"
	"fmt"
	"google.golang.org/protobuf/proto"
	"strings"
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
	_ = s.AddHandler(pb.MessageId_ChatResponse, s.onMessage)
	_ = s.AddHandler(pb.MessageId_LeaveChannelResponse, s.onLeaveChannel)
	_ = s.AddHandler(pb.MessageId_UserActionNotify, s.onUserAction)
	console.NewConsole().AddHandler("chat", s.cmdChat)
	console.NewConsole().AddHandler("leave", s.cmdLeaveChannel)
	logger.Info("ENTER CHANNEL")
}

func (s *SessionStateChannel) OnExit() {
	s.DelHandler(pb.MessageId_ChatResponse)
	s.DelHandler(pb.MessageId_LeaveChannelResponse)
	s.DelHandler(pb.MessageId_UserActionNotify)
	console.NewConsole().DelHandler("chat")
	console.NewConsole().DelHandler("leave")
	logger.Info("LEAVE CHANNEL")
}

func (s *SessionStateChannel) cmdChat(params []string) {
	if 0 == len(params) {
		logger.Error("no chat content")
		return
	}
	req := &pb.ChatRequestMessage{
		Message: strings.Join(params, " "),
	}
	s.SendMessage(pb.MessageId_ChatRequest, req)
}

func (s *SessionStateChannel) onMessage(_ uint32, data []byte) error {
	resp := &pb.ChatResponseMessage{}
	if err := proto.Unmarshal(data, resp); nil != err {
		return err
	}
	
	fmt.Printf("%s says: %s.\n", resp.Username, resp.Message)
	return nil
}

func (s *SessionStateChannel) cmdLeaveChannel([]string) {
	req := &pb.LeaveChannelRequestMessage{}
	s.SendMessage(pb.MessageId_LeaveChannelRequest, req)
}

func (s *SessionStateChannel) onLeaveChannel(_ uint32, data []byte) error {
	resp := &pb.LeaveChannelResponseMessage{}
	if err := proto.Unmarshal(data, resp); nil != err {
		return err
	}
	if resp.Result == pb.Result_Success {
		s.GetSession().Translate("Lobby")
	}
	return nil
}

func (s *SessionStateChannel) onUserAction(_ uint32, data []byte) error {
	resp := &pb.UserActionNotifyMessage{}
	if err := proto.Unmarshal(data, resp); nil != err {
		return err
	}
	switch resp.Type {
	case pb.UserActionType_EnterChannel:
		logger.Info("user %s enter channel.", resp.Username)
		break
	case pb.UserActionType_LeaveChannel:
		logger.Info("user %s leave channel.", resp.Username)
		break
	}
	return nil
}

