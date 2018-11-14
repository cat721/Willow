package Message

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
)

type Message struct {
	Header
	Payload []byte
}

func (m *Message) Deserialize(r io.Reader) error {
	err := m.Header.Deserialize(r)
	if err != nil{
		return err
	}

	m.Payload,err = ioutil.ReadAll(r)

	if err != nil{
		return err
	}

	return nil
}

func (m *Message) Serialize (w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil{
		return err
	}

	_,err = w.Write(m.Payload)

	if err != nil{
		return err
	}

	return nil
}

func NewMessage(msgType uint32,payload []byte) *Message {


	h := Header{
		MsgType:msgType,
	}

	m := Message{
		Header:h,
		Payload:payload,
	}

	hash := sha256.Sum256(m.Payload)

	copy(m.Header.Checksum[:],hash[:4])
	return &m
}