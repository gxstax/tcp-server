package packet

import (
	"bytes"
	"fmt"
	"sync"
)

const (
	CommandConn = iota + 0x01	// 0x01，连接请求包
	CommandSubmit				// 0x02，消息请求包
)

const (
	CommandConnAck = iota + 0x81	// 0x81，连接请求响应包
	CommandSubmitAck				// 0x82，消息请求响应包
)

type Packet interface {
	Encode() ([]byte, error)	// []byte -> struct
	Decode([]byte) error		// struct -> []byte
}

type Submit struct {
	ID		string
	Payload	[]byte
}

func (s *Submit) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8])
	s.Payload = pktBody[8:]
	return nil
}

func (s *Submit) Encode() ([]byte, error) {
	pakBytes := bytes.Join([][]byte{[]byte(s.ID[:8]), s.Payload}, nil)
	return pakBytes, nil
}

type SubmitAck struct {
	ID 		string
	Result 	uint8
}

func (s *SubmitAck) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8])
	s.Result = uint8(pktBody[8])
	return nil
}

func (s *SubmitAck) Encode() ([]byte, error) {
	join := bytes.Join([][]byte{[]byte(s.ID[:8]), []byte{s.Result}}, nil)
	return join, nil
}

var SubmitPool = sync.Pool{
	New: func() interface{} {
		return &Submit{}
	},
}

func Decode(packet []byte) (Packet, error) {
	commandID := packet[0]
	pktBody := packet[1:]

	switch commandID {
	case CommandConn:
		return nil, nil
	case CommandConnAck:
		return nil, nil
	case CommandSubmit:
		//s := Submit{}
		s := SubmitPool.Get().(*Submit) // 从 SubmitPool 池中获取一个 Submit 内存对象
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return s, nil
	case CommandSubmitAck:
		s := SubmitAck{}
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return &s, nil
	default:
		return nil, fmt.Errorf("unknown commandID[%d]", commandID)
	}
}

func Encode(p Packet) ([]byte, error) {
	var commandID uint8
	var pktBody []byte
	var err error

	switch t := p.(type) {
	case *Submit:
		commandID = CommandSubmit
		pktBody, err = t.Encode()
		if err != nil {
			return nil, err
		}
	case *SubmitAck:
		commandID = CommandSubmitAck
		pktBody, err = t.Encode()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type [%s]", t)
	}

	return bytes.Join([][]byte{[]byte{commandID}, []byte(pktBody)}, nil), nil
}