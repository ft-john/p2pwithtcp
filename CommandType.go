package main

type CommandType byte

const (
	Heartbeat CommandType = 0x00
	GetAddr   CommandType = 0x01
	SendAddr  CommandType = 0x02
	CommonMsg CommandType = 0x03
)

func GetCommandName(cmdType CommandType) string {
	switch cmdType {
	case 0x00:
		return "Heartbeat"
	case 0x01:
		return "GetAddr"
	case 0x2:
		return "SendAddr"
	case 0x03:
		return "CommonMsg"
	default:
		return ""
	}
}
