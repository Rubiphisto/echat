package sessions

import (
	"echat/common/pb"
	"google.golang.org/protobuf/proto"
	"time"
)

const LATEST_MSG_COUNT = 50

type ChatMessage struct {
	msgNo		uint32
	contents	*pb.ChatContent
}

type Channel struct {
	name		string
	users		map[string]time.Time
	latestMsg	[LATEST_MSG_COUNT]*ChatMessage
	msgNo		uint32
}

func NewChannel(channelName string) *Channel {
	return &Channel{
		name:  channelName,
		users: make(map[string]time.Time),
	}
}

func (c *Channel) AddUser(user *User) {
	_, ok := c.users[user.GetUserName()]
	if ok { // 已经加入房间
		return
	}
	notify := &pb.UserActionNotifyMessage{
		Type:     pb.UserActionType_EnterChannel,
		Username: user.GetUserName(),
	}
	c.Broadcast(pb.MessageId_UserActionNotify, notify)
	
	c.users[user.GetUserName()] = time.Now()
	resp := &pb.EnterChannelResponseMessage{
		ChannelName: c.name,
		Users:       nil,
		//Messages:    &c.latestMsg,
	}
	if c.msgNo > LATEST_MSG_COUNT {
		for i := uint32(0); i < LATEST_MSG_COUNT; i++ {
			index := (c.msgNo + i) % LATEST_MSG_COUNT
			resp.Contents = append(resp.Contents, c.latestMsg[index].contents)
		}
	} else {
		for i := uint32(0); i < c.msgNo; i++ {
			resp.Contents = append(resp.Contents, c.latestMsg[i].contents)
		}
	}
	for username, _ := range c.users {
		resp.Users = append(resp.Users, username)
	}
	user.OnEnterChannel(c.name)
	user.SendMessage(pb.MessageId_EnterChannelResponse, resp)
}

func (c *Channel) DelUser(user *User) {
	_, ok := c.users[user.GetUserName()]
	if !ok {
		return
	}
	delete(c.users, user.GetUserName())

	notify := &pb.UserActionNotifyMessage{
		Type:     pb.UserActionType_LeaveChannel,
		Username: user.GetUserName(),
	}
	c.Broadcast(pb.MessageId_UserActionNotify, notify)
	
	user.OnLeaveChannel()
}

func (c *Channel) Chat(username string, words string) {
	_, ok := c.users[username]
	if !ok {
		return
	}
	
	// TODO: filter the dirty word
	
	index := c.msgNo % LATEST_MSG_COUNT
	c.latestMsg[index] = &ChatMessage{
		msgNo: c.msgNo,
		contents: &pb.ChatContent{
			User:  username,
			Words: words,
		},
	}
	c.msgNo++
	
	msg := &pb.ChatResponseMessage{
		Username: username,
		Message:  words,
	}
	c.Broadcast(pb.MessageId_ChatResponse, msg)
}

func (c *Channel) Broadcast(msgId pb.MessageId, message proto.Message) {
	for username, _ := range c.users {
		user := GetUserManager().GetUser(username)
		if nil == user {
			continue
		}
		user.SendMessage(msgId, message)
	}
}
