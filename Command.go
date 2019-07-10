package main

import (
	"crypto/sha256"
	"encoding/binary"
	//"log"
)

const (
	Prefix  uint16 = 0x7788
	Suffix  uint16 = 0x9900
	Prefix1 byte   = 0x77
	Prefix2 byte   = 0x88
	Suffix1 byte   = 0x99
	Suffix2 byte   = 0x00
)

type Command struct {
	Prefix  uint16
	ID      string //GUID: 32 bytes
	Length  uint32
	CmdType CommandType
	Message string
	CRC     uint16
	Suffix  uint16
}

func NewCommand(id string, cmdType CommandType, msg string) Command {
	cmd := Command{}
	cmd.Prefix = Prefix
	cmd.ID = id
	cmd.Length = uint32(1 + len([]byte(msg)))
	cmd.CmdType = cmdType
	cmd.Message = msg

	bufData := make([]byte, 4)
	binary.BigEndian.PutUint32(bufData, cmd.Length)
	bufData = append(bufData, byte(cmd.CmdType))
	bufData = append(bufData, []byte(cmd.Message)...)
	//log.Printf("CRC Len: %d\n", len(bufData))
	bufCrc := CalculateCrc(bufData)
	cmd.CRC = binary.BigEndian.Uint16(bufCrc)

	cmd.Suffix = Suffix
	return cmd
}

func (cmd Command) UnMarshal() []byte {
	var buffer, bufMsg, data []byte
	var bufPrefix = make([]byte, 2)
	var bufLength = make([]byte, 4)
	var bufCrc = make([]byte, 2)
	var bufSuffix = make([]byte, 2)

	bufMsg = []byte(cmd.Message)
	// Data length:
	// CmdType: 1
	// Message:
	length := 1 + len(bufMsg)

	binary.BigEndian.PutUint16(bufPrefix, Prefix)
	buffer = append(buffer, bufPrefix...)

	buffer = append(buffer, []byte(cmd.ID)...)

	binary.BigEndian.PutUint32(bufLength, uint32(length))
	buffer = append(buffer, bufLength...)

	buffer = append(buffer, byte(cmd.CmdType))
	buffer = append(buffer, bufMsg...)

	//Calculate crc
	data = append(bufLength, byte(cmd.CmdType))
	data = append(data, bufMsg...)
	bufCrc = CalculateCrc(data)
	buffer = append(buffer, bufCrc...)

	binary.BigEndian.PutUint16(bufSuffix, Suffix)
	buffer = append(buffer, bufSuffix...)

	//fmt.Println(buffer)
	return buffer
}

func CalculateCrc(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:2]
}
