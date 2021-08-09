package minecraft

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"phoenix/minecraft/auth"
	"phoenix/minecraft/function"
	"phoenix/minecraft/ligo"
	"phoenix/minecraft/protocol/packet"
)

const Operator = "CAIMEOX"
const maxWorkers = 100

type Worker struct {
	spaces map[string]*function.Space
	virtualMachine *ligo.VM
}

func (w *Worker) GetSpace(name string) *function.Space {
	return w.spaces[name]
}

func Run(address string) {
	pterm.EnableDebugMessages()
	worker := Worker{
		spaces:    make(map[string]*function.Space),
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

	worker.virtualMachine.Funcs["plot"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
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


	dialer := Dialer{
		TokenSource: auth.TokenSource,
	}
	conn, err := dialer.Dial("raknet", address)
	if err != nil {
		pterm.Error.Println(err)
	}
	defer conn.Close()

	worker.virtualMachine.Funcs["get"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		err := conn.SendCommand("testforblock ~ ~ ~ air", func(output *packet.CommandOutput) error {
			pos, _ := function.SliceAtoi(output.OutputMessages[0].Parameters)
			if len(pos) != 3 {
				return errors.New("testforblock function have got wrong number of positions")
			} else {
				space := worker.GetSpace(vm.Vars["space"].Value.(string))
				space.SetPointer(pos)
				pterm.Info.Println(fmt.Sprintf("Position got: %v", pos))
			}
			return nil
		})
		if err != nil {
			return vm.Throw(fmt.Sprintf("SendCommand: %s", err))
		} else {
			return ligo.Variable{Type: ligo.TypeNil, Value: nil}
		}
	}

	if err := conn.DoSpawn(); err == nil {
		// Collector : Get Position
		eval, err := worker.virtualMachine.Eval(`(get)`)
		if err != nil {
			pterm.Info.Println(err)
		} else if eval.Value != nil {
			pterm.Info.Println(eval.Value)
		}
	} else {
		pterm.Error.Println(err)
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
			if p.TextType == packet.TextTypeChat {
				if Operator == p.SourceName {
					pterm.Info.Println(fmt.Sprintf("[%s] %s", p.SourceName, p.Message))
					value, err := worker.virtualMachine.Eval(p.Message)
					if err != nil {
						pterm.Error.Println(err)
					} else {
						pterm.Info.Println(value.Value)
					}
				}
			}

		case *packet.CommandOutput:
			callback, ok := conn.callbacks[p.CommandOrigin.UUID.String()]
			delete(conn.callbacks, p.CommandOrigin.UUID.String())
			// TODO : Handle !ok
			if ok {
				if !p.OutputMessages[0].Success {
					pterm.Warning.Println(fmt.Sprintf("Unknown command: %s. Please check that the command exists and that you have permission to use it.", p.OutputMessages[0].Parameters[0] ))
				}
				err := callback(p)
				if err != nil {
					pterm.Warning.Println(err)
					// TODO : Handle error
				}
				continue
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

