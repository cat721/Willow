package Message

import (
	"encoding/binary"
	"io"
)

const CheckSumSize = 4

type Header struct {
	MsgType uint32
	Checksum [CheckSumSize]byte
}

func (h *Header)Deserialize(r io.Reader) error {
	err:=binary.Read(r,binary.LittleEndian,&h.MsgType)
	if err != nil {
		return err
	}

	err = binary.Read(r,binary.LittleEndian,h.Checksum[:])
	if err != nil {
		return err
	}

	return nil
}

func (h *Header) Serialize (w io.Writer) error {
	err := binary.Write(w,binary.LittleEndian,h.MsgType)
	if err != nil {
		return err
	}

	err = binary.Write(w,binary.LittleEndian,h.Checksum[:])
	if err != nil {
		return err
	}

	return nil
}
