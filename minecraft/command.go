package minecraft

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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