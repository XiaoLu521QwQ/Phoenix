package minecraft

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"phoenix/minecraft/auth"
	"phoenix/minecraft/function"
	"phoenix/minecraft/ligo"
	"phoenix/minecraft/protocol"
	"phoenix/minecraft/protocol/packet"
)

const Operator = "CAIMEOX"
const maxWorkers = 100
type Callback func(output *packet.CommandOutput) error

type Worker struct {
	spaces map[string]*function.Space
	callbacks map[string]Callback
	virtualMachine *ligo.VM
}

func Run(address string) {
	worker := Worker{
		spaces:    make(map[string]*function.Space),
		callbacks: make(map[string]Callback),
		virtualMachine: ligo.NewVM(),
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}
	worker.virtualMachine.Vars["space"] = ligo.Variable{
		Type:  ligo.TypeString,
		Value: "overworld",
	}

	worker.spaces["overworld"] = function.NewSpace()

	var Plot ligo.InBuilt
	Plot = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		work_space := worker.spaces[vm.Vars["space"].Value.(string)]
		if variable[0].Type == ligo.TypeArray {
			work_space.PlotArray(variable[0].Value.([]function.Vector))
		} else if variable[0].Type == ligo.TypeFloat {
			work_space.Plot(variable[0].Value.(function.Vector))
		} else {
			return ligo.Variable{Type: ligo.TypeErr, Value: "plot function's first argument should be of a vector or vector slice type"}
		}
		return ligo.Variable {
			Type: ligo.TypeNil,
		}
	}
	worker.virtualMachine.Funcs["plot"] = Plot


	dialer := Dialer{
		TokenSource: auth.TokenSource,
	}
	conn, err := dialer.Dial("raknet", address)
	if err != nil {
		pterm.Error.Println(err)
	}
	defer conn.Close()
	worker.virtualMachine.Funcs["get"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		err := SendCommand("testforblock ~ ~ ~ air", func(output *packet.CommandOutput) error {
			pos, _ := function.SliceAtoi(output.OutputMessages[0].Parameters)
			if len(pos) != 3 {
				return errors.New("testforblock function have got wrong number of positions")
			} else {
				pterm.Info.Println(fmt.Sprintf("Position got: %v", pos))
			}
			return nil
		}, conn, worker.callbacks)
		if err != nil {
			return vm.Throw(fmt.Sprintf("SendCommand: %s", err))
		} else {
			return ligo.Variable{Type: ligo.TypeNil, Value: nil}
		}
	}
	if err := conn.DoSpawn(); err == nil {
		eval, err := worker.virtualMachine.Eval(`(get)`)
		if err != nil {
			pterm.Info.Println(err)
			//panic(err)
		}
		pterm.Info.Println(eval.Value)
	} else {
		pterm.Info.Println(err)
	}

	// You will then want to start a for loop that reads packets from the connection until it is closed.
	for {
		// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
		// a read timeout is set. You will generally want to return or break if this happens.
		pk, err := conn.ReadPacket()
		if err != nil {
			break
		}

		// The pk variable is of type packet.Packet, which may be type asserted to gain access to the data
		// they hold:
		switch p := pk.(type) {
		case *packet.Text:
			if Operator == p.SourceName {
				pterm.Info.Println(fmt.Sprintf("[%s] %s", p.SourceName, p.Message))
			}
		case *packet.CommandOutput:
			callback, ok := worker.callbacks[p.CommandOrigin.UUID.String()]
			delete(worker.callbacks, p.CommandOrigin.UUID.String())
			// TODO : Handle !ok
			if ok {
				err := callback(p)
				if err != nil {
					// TODO : Handle error
					return
				}
			}
		}


		// Write a packet to the connection: Similarly to ReadPacket, WritePacket will (only) return an error
		// if the connection is closed.
		p := &packet.RequestChunkRadius{ChunkRadius: 32}
		if err := conn.WritePacket(p); err != nil {
			break
		}
	}
}

func SendCommand(command string, callback Callback, conn *Conn, cbMap map[string]Callback) error {
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
	cbMap[callbackID.String()] = callback
	return conn.WritePacket(commandRequest)
}