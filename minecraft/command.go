package minecraft

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"phoenix/minecraft/function"
	"phoenix/minecraft/protocol"
	"phoenix/minecraft/protocol/packet"
	"time"
)

func (conn *Conn) SendCommand(command string, callback Callback) error {
	requestID := uuid.New()
	callbackID := uuid.New()
	commandRequest := &packet.CommandRequest{
		CommandOrigin: protocol.CommandOrigin{
			Origin:         protocol.CommandOriginPlayer,
			UUID:           callbackID,
			RequestID:      requestID.String(),
			PlayerUniqueID: 0,
		},
		CommandLine: command,
		Internal: false,
	}
	conn.callbacks[callbackID.String()] = callback
	return conn.WritePacket(commandRequest)
}

func (conn *Conn) SendCommandWO(command string) error {
	commandRequest := &packet.SettingsCommand{
		CommandLine:    command,
		SuppressOutput: false,
	}
	return conn.WritePacket(commandRequest)
}

func (conn *Conn) SendCommandNoCallback(command string) error {
	requestID := uuid.New()
	callbackID := uuid.New()
	commandRequest := &packet.CommandRequest{
		CommandOrigin: protocol.CommandOrigin{
			Origin:         protocol.CommandOriginPlayer,
			UUID:           callbackID,
			RequestID:      requestID.String(),
			PlayerUniqueID: 0,
		},
		CommandLine: command,
		Internal: false,
	}
	return conn.WritePacket(commandRequest)
}

// BarWriter Stdout
type BarWriter struct {
	conn *Conn
}

func (b BarWriter) Write(p []byte) (int, error) {
	err := b.conn.Actionbar(b.conn.worldConfig.operator, string(p))
	return 0, err
}

func (conn *Conn) Actionbar(target, text string) error {
	return conn.SendCommandNoCallback(fmt.Sprintf("title %s actionbar %s", target, text))
}

func (conn *Conn) SetBlock(pos function.Vector) error {
	cmd := fmt.Sprintf("setblock %v %v %v %s %d", pos[0], pos[1], pos[2], conn.worldConfig.block.name, conn.worldConfig.block.data)
	return conn.SendCommandNoCallback(cmd)
}

func (conn *Conn) Info(text ...string) error {
	return conn.SendCommand(InfoRequest("@a", text...), func(output *packet.CommandOutput) error {return nil})
}

func (conn *Conn) Error(text ...string) error {
	return conn.SendCommand(ErrorRequest("@a", text...), func(output *packet.CommandOutput) error {return nil})
}


func InfoRequest(target string, lines ...string) string {
	now := time.Now().Format("§6[15:04:05]§b INFO: ")
	var items []TellrawItem
	for _, text := range lines {
		msg := fmt.Sprintf("%v %v", now, text)
		items=append(items,TellrawItem{Text:msg})
	}
	final := &TellrawStruct {
		RawText: items,
	}
	content, _ := json.Marshal(final)
	cmd := fmt.Sprintf("tellraw %v %s", target, content)
	return cmd
}

func ErrorRequest(target string, lines ...string) string {
	now := time.Now().Format("§6[15:04:05]§c ERROR: ")
	var items []TellrawItem
	for _, text := range lines {
		msg := fmt.Sprintf("%v %v", now, text)
		items = append(items,TellrawItem{Text:msg})
	}
	final := &TellrawStruct {
		RawText: items,
	}
	content, _ := json.Marshal(final)
	cmd := fmt.Sprintf("tellraw %v %s", target, content)
	return cmd
}

type TellrawItem struct {
	Text string `json:"text"`
}

type TellrawStruct struct {
	RawText []TellrawItem `json:"rawtext"`
}