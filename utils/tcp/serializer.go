package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ConnectSerializer 网络连接的数据编码器
type ConnectSerializer interface {
	// Serialize 将数据编码并写入网络
	Serialize(myID uint32, writer io.Writer, content []byte) error
}

// ConnectDeserializer 网络连接的数据解码器
type ConnectDeserializer interface {
	// Deserialize 从网络读入数据并解码
	Deserialize(myID uint32, reader io.Reader) ([]byte, error)
}

// SerializeFactory 序列化工厂
// 编码器和解码器在不同的goroutine执行
type SerializeFactory interface {
	CreateSerializer() ConnectSerializer
	CreateDeserializer() ConnectDeserializer
}

// GetDefaultSerializeFactory 获得默认的序列化工厂
// 以4字节来保存整个包(含4字节本身)长度，长度以order顺序写入到缓存中
// 单个网络包最大长度不超过5M
func GetDefaultSerializeFactory(order binary.ByteOrder) SerializeFactory {
	ist.order = order
	return &ist
}

type defaultSerializeFactory struct {
	order binary.ByteOrder
}

var ist defaultSerializeFactory

// CreateSerializer 序列化器
func (f *defaultSerializeFactory) CreateSerializer() ConnectSerializer {
	return f
}
// CreateDeserializer 反序列化器
func (f *defaultSerializeFactory) CreateDeserializer() ConnectDeserializer {
	return f
}

const (
	wholeHeadSize = 4               // 整个消息包头长度
	msgMax        = 5 * 1024 * 1024 // 消息最大长度
)

func (f *defaultSerializeFactory) Deserialize(myID uint32, reader io.Reader) ([]byte, error) {
	head := make([]byte, wholeHeadSize)
	if _, err := io.ReadFull(reader, head); nil != err {
		return nil, err
	}
	packetLength := f.order.Uint32(head)
	if packetLength > msgMax {
		return nil, fmt.Errorf("PacketSerializer read pack size %v is greater than max length %v", packetLength, msgMax)
	}
	msgLength := packetLength - wholeHeadSize

	msg := make([]byte, msgLength)
	if _, err := io.ReadFull(reader, msg); nil != err {
		return nil, err
	}
	return msg, nil
}

func (f *defaultSerializeFactory) Serialize(myID uint32, writer io.Writer, content []byte) error {
	head := make([]byte, wholeHeadSize)

	f.order.PutUint32(head, uint32(len(content)+wholeHeadSize))
	_, err := writer.Write(head)
	if nil != err {
		return err
	}
	_, err = writer.Write(content)
	return err
}

