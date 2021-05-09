package tcp

// Session 网络会话对象，由 SessionFactory.CreateSession创建
// 所有回调均由网络层创建的同一 goroutine 发起
type Session interface {
	// Initialize 连接建立后被调用
	Initialize(connection Connection) error
	// Uninitialized 连接关闭后被调用
	Uninitialized()
	// OnRecvMessage 收到数据包
	OnRecvMessage(content []byte)
	// CheckHeartbeat 心跳检测，返回 false 表示断开网络连接
	CheckHeartbeat() bool
}

// SessionFactory 网络会话对象创建工厂
type SessionFactory interface {
	// CreateSession 创建网络会话对象
	CreateSession() Session
}

