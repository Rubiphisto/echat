package sessions

import (
	"echat/common/pb"
	"echat/utils/logger"
	"google.golang.org/protobuf/proto"
)

type User struct {
	userName			string
	session				*Session
	channelName			string
}

func (u *User) GetUserName() string {
	return u.userName
}

func (u *User) IsInChannel() bool {
	return 0 != len(u.channelName)
}

func (u *User) GetChannelName() string {
	return u.channelName
}

func (u *User) SendMessage(msgId pb.MessageId, message proto.Message) {
	if nil == u.session {
		return
	}
	u.session.SendMessage(uint32(msgId), message)
}

func (u *User) LeavelChannel() {
	if 0 == len(u.channelName) {
		return
	}
	channel := GetChannelManager().GetChannel(u.channelName)
	if nil == channel {
		return
	}
	channel.DelUser(u)
}

func (u *User) OnEnterChannel(channelName string) {
	u.channelName = channelName
	logger.Info("User %v enter channels %v", u.userName, u.channelName)
}

func (u *User) OnLeaveChannel() {
	logger.Info("User %v leave channels %v", u.userName, u.channelName)
	u.channelName = ""
	resp := &pb.LeaveChannelResponseMessage{Result: pb.Result_Success}
	u.SendMessage(pb.MessageId_LeaveChannelResponse, resp)
}
