package pack

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	// ServerByteOrder 服务端使用的字节序
	ServerByteOrder = binary.LittleEndian
)

const (
	// PerHeadSize 单个数据包的包头字节数 sizeof(uint16)
	PerHeadSize = 2

	// MsgIDSize 消息号的字节数 sizeof(uint32)
	MsgIDSize = 4

	// WholeMsgMax 整个数据包的最大字节数(1MB)
	WholeMsgMax = 1024 * 1024
)

// MsgPack 单个包的结构，元素的顺序与二进制消息的顺序一致
type MsgPack struct {
	Length uint16 // Length Data的字节数
	MsgId  uint32 // MsgId 消息编号
	Data   []byte // Data 真实数据
}

func Pack(pack *MsgPack) ([]byte, error) {
	wholeLen := getPackLength(pack)
	if wholeLen > WholeMsgMax {
		return nil, fmt.Errorf("PackEx data length overflow: %d > %d", wholeLen, WholeMsgMax)
	}

	//if 0 == len(pack.Data) {
	//	return nil, fmt.Errorf("the data of message %v length is zero", pack.MsgId)
	//}

	buff := bytes.NewBuffer(make([]byte, 0, wholeLen))
	binary.Write(buff, ServerByteOrder, uint16(len(pack.Data)))
	binary.Write(buff, ServerByteOrder, pack.MsgId)
	binary.Write(buff, ServerByteOrder, pack.Data)

	return buff.Bytes(), nil
}

// Unpack 解包二进制数据内容
func Unpack(data []byte) (*MsgPack, error) {
	cursor := 0
	dataLength := int(ServerByteOrder.Uint16(data[cursor : cursor+PerHeadSize]))
	cursor += PerHeadSize
	msgID := ServerByteOrder.Uint32(data[cursor : cursor+MsgIDSize])
	cursor += MsgIDSize

	if len(data) < PerHeadSize + MsgIDSize + dataLength {
		return nil, fmt.Errorf("pack length is invalid")
	}

	return &MsgPack{
		Length: uint16(dataLength),
		MsgId:  msgID,
		Data:   data[cursor : cursor+ dataLength],
	}, nil
}

// 计算MsgPack对应的整包字节数
func getPackLength(pack *MsgPack) int {
	return len(pack.Data) + PerHeadSize + MsgIDSize
}

