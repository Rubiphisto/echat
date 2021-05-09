package sessions

var (
	channelManager ChannelManager
)

type ChannelManager struct {
	channels map[string]*Channel
}

func init() {
	channelManager.channels = make(map[string]*Channel)
}

func GetChannelManager() *ChannelManager {
	return &channelManager
}

func (m *ChannelManager) EnterChannel(user *User, channelName string) *Channel {
	channel, ok := m.channels[channelName]
	if !ok {
		channel = NewChannel(channelName)
		if nil == channel {
			return nil
		}
		m.channels[channel.name] = channel
	}
	channel.AddUser(user)
	return channel
}

func (m *ChannelManager) GetChannel(channelName string) *Channel {
	channel, ok := m.channels[channelName]
	if !ok {
		return nil
	}
	return channel
}
