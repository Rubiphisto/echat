package session

import (
	"echat/client/console"
	"echat/common/pb"
	"echat/utils/logger"
	"fmt"
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
	_ = s.AddHandler(pb.MessageId_EnterChannelResponse, s.onEnterChannel)
	console.NewConsole().AddHandler("enter", s.cmdEnterChannel)
	logger.Info("ENTER LOBBY")
}

func (s *SessionStateLobby) OnExit() {
	s.DelHandler(pb.MessageId_EnterChannelResponse)
	console.NewConsole().DelHandler("enter")
	logger.Info("LEAVE LOBBY")
}

func (s *SessionStateLobby) cmdEnterChannel(params []string) {
	if 0 == len(params) {
		logger.Error("no channel name")
		return
	}
	req := &pb.EnterChannelRequestMessage{
		ChannelName: params[0],
	}
	s.SendMessage(pb.MessageId_EnterChannelRequest, req)
}

func (s *SessionStateLobby) onEnterChannel(_ uint32, data []byte) error {
	resp := &pb.EnterChannelResponseMessage{}
	if err := proto.Unmarshal(data, resp); nil != err {
		return err
	}
	fmt.Printf("enter channel %v with result %v\n", resp.ChannelName, resp.Result)
	if pb.Result_Success == resp.Result {
		fmt.Printf("enter channel [%v] and there are %d user\n", resp.ChannelName, len(resp.Users))
		for _, content := range resp.Contents {
			fmt.Printf("%s Says: %s.\n", content.User, content.Words)
		}
		s.GetSession().Translate("Channel")
	}
	return nil
}

