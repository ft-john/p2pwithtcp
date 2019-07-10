package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	//"Fiii/p2pWithTcp/cmd/p2pWithTcp"

	reuse "github.com/libp2p/go-reuseport"
	tsgutils "github.com/typa01/go-utils"
)

type Cmd struct {
	Conn   net.Conn
	Writer *bufio.Writer
	Data   Command
}

type Peer struct {
	Address string
	Conn    net.Conn
	ID      string
	Reader  *bufio.Reader
	Writer  *bufio.Writer
}

const (
	TRACKER_ADDRESS = "127.0.0.1:8880"
)

var (
	wg           sync.WaitGroup
	server       net.Listener
	err          error
	cmdQueue     chan Cmd
	tracker      *bool
	port         *int
	localAddress string
	peers        map[string]*Peer
	nodeID       string
)

func startListen() {
	defer server.Close()

	for {
		conn, err := server.Accept()

		if err != nil {
			log.Printf("Error in appcept: %s\n", err.Error())
			continue
		} else {
			log.Printf("New peer connected:%s\n", conn.RemoteAddr().String())
			address := conn.RemoteAddr().String()
			peers[address] = &Peer{
				Address: address,
				Conn:    conn,
				Reader:  bufio.NewReader(conn),
				Writer:  bufio.NewWriter(conn),
			}
			//wg.Add(1)
			go readData(peers[address])
		}
	}

	//wg.Wait()
}

func connect(remotAddress string) {
	conn, err := reuse.Dial("tcp", localAddress, remotAddress)
	if err != nil {
		log.Println(err)
	} else {
		address := conn.RemoteAddr().String()
		peers[address] = &Peer{
			Address: address,
			Conn:    conn,
			Reader:  bufio.NewReader(conn),
			Writer:  bufio.NewWriter(conn),
		}

		go readData(peers[address])

		if remotAddress == TRACKER_ADDRESS {
			cmdQueue <- Cmd{conn, peers[address].Writer, NewCommand(nodeID, GetAddr, "")}
		}
	}

}

func disConnectByPeer(peer *Peer) {
	fmt.Println("Close client:", peer.Address)
	delete(peers, peer.Address)
	peer.Conn.Close()
}

func disConnectByConn(conn net.Conn) {
	peer, ok := peers[conn.RemoteAddr().String()]
	if ok {
		disConnectByPeer(peer)
	}
}

func readData(peer *Peer) {
	defer disConnectByPeer(peer)
	state := 0x00
	var buflen []byte
	var bufcrc []byte
	var bufdata []byte
	var bufID []byte
	//var newCmd Command
	var length uint32
	var cmdType byte
	var cursor uint32

	for {
		recvByte, err := peer.Reader.ReadByte()
		if err != nil {
			log.Println(err.Error())
			//log.Println("Client disconnected")

			break
		} /*else if data != "" && data != "\n" {
			cmdQueue <- Cmd{peer.Conn, data}
			log.Println("Received Data:", data)
		}*/
		//log.Printf("state: 0x%02X, byte:0x%02X\n", state, recvByte)

		switch state {
		case 0x00:
			if recvByte == Prefix1 {
				state = 0x01
			} else {
				state = 0x00
			}
			break
		case 0x01:
			if recvByte == Prefix2 {
				state = 0x02
				//initialize
				buflen = make([]byte, 4)
				bufcrc = make([]byte, 2)
				bufdata = make([]byte, 0)
				bufID = make([]byte, 32)
				cursor = 0
			} else {
				state = 0x00
			}
			break
		case 0x02:
			bufID[cursor] = recvByte
			cursor++
			if cursor == 32 {
				state = 0x03
				cursor = 0
			}
		case 0x03:
			buflen[cursor] = recvByte
			cursor++

			if cursor == 4 {
				state = 0x04
				cursor = 0
				length = binary.BigEndian.Uint32(buflen)
			}
			break
		case 0x04:
			cmdType = recvByte
			if length <= 1 {
				state = 0x06
			} else {
				state = 0x05
			}
			break
		case 0x05:
			bufdata = append(bufdata, recvByte)
			cursor++
			if cursor == length-1 {
				state = 0x06
				cursor = 0
			}
			break
		case 0x06:
			bufcrc[cursor] = recvByte
			cursor++

			if cursor == 2 {
				crcData := append(buflen, cmdType)
				crcData = append(crcData, bufdata...)
				crc := CalculateCrc(crcData)

				//log.Println("CRC Data len", len(crcData))

				if crc[0] == bufcrc[0] && crc[1] == bufcrc[1] {
					state = 0x07
				} else {
					log.Println("CRC Error")
					state = 0x00
				}
			}
			break
		case 0x07:
			if recvByte == Suffix1 {
				state = 0x08
			} else {
				state = 0x00
			}
			break
		case 0x08:
			if recvByte == Suffix2 {

				if peer.ID == "" {
					peer.ID = string(bufID)
				}

				HandleCommand(peer, NewCommand(peer.ID, CommandType(cmdType), string(bufdata)))
			}
			state = 0x00
			break
		}
	}
}

func HandleCommand(peer *Peer, cmd Command) {
	log.Printf("Received new command from %s: %s, data: %s", peer.Address, GetCommandName(cmd.CmdType), cmd.Message)

	switch cmd.CmdType {
	case GetAddr:
		data := ""
		for _, p := range peers {
			if data != "" {
				data += "|"
			}

			data += p.Address + "," + p.ID
		}

		newCmd := NewCommand(nodeID, SendAddr, data)
		cmdQueue <- Cmd{peer.Conn, peer.Writer, newCmd}
		break
	case SendAddr:
		for _, item := range strings.Split(cmd.Message, "|") {
			strs := strings.Split(item, ",")
			if len(strs) == 2 && strs[1] != nodeID {
				_, ok := peers[strs[0]]
				if !ok {
					log.Println("Connect to new peer: ", strs[0])
					go connect(strs[0])
				}
			}
		}
		break
	case Heartbeat:
		break
	}
}

func writeData() {
	for {
		msg := <-cmdQueue
		data := msg.Data.UnMarshal()
		_, err := msg.Writer.Write(data)
		if err != nil {
			log.Println(err)
			disConnectByConn(msg.Conn)
		}

		err = msg.Writer.Flush()
		if err != nil {
			log.Println(err)
			disConnectByConn(msg.Conn)
		} else {
			log.Printf("Send data to %s: %s", msg.Conn.RemoteAddr().String(), GetCommandName(msg.Data.CmdType))
		}
	}
}

func sendHeartbeat() {
	for {
		for _, peer := range peers {
			cmdQueue <- Cmd{
				Conn:   peer.Conn,
				Writer: peer.Writer,
				Data:   NewCommand(peer.ID, Heartbeat, ""),
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {
	tracker = flag.Bool("tracker", false, "Run as a tracker server")
	port = flag.Int("port", 8881, "Source port number")
	flag.Parse()

	cmdQueue = make(chan Cmd, math.MaxInt16)
	peers = make(map[string]*Peer)
	nodeID = tsgutils.GUID()

	localAddress = ":" + strconv.Itoa(*port)
	if *tracker {
		localAddress = TRACKER_ADDRESS
	}

	server, err = reuse.Listen("tcp", localAddress)
	if err != nil {
		panic(err)
	}

	log.Println("Node started, waiting for connections ...")
	go writeData()

	if !*tracker {
		go connect(TRACKER_ADDRESS)
	}

	go sendHeartbeat()
	startListen()
}
