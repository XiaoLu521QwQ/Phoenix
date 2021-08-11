package minecraft

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/schollz/progressbar/v3"
	"phoenix/minecraft/auth"
	"phoenix/minecraft/function"
	"phoenix/minecraft/function/generator"
	"phoenix/minecraft/ligo"
	"phoenix/minecraft/protocol/packet"
	"time"
)


type Worker struct {
	spaces map[string]*function.Space
	virtualMachine *ligo.VM
}

func (w *Worker) GetSpace(name string) *function.Space {
	return w.spaces[name]
}

func Run(path string) {
	config := function.ReadConfig(path)
	if config.Debug.Enabled {
		pterm.EnableDebugMessages()
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}
	worker := Worker{
		spaces:    make(map[string]*function.Space),
		virtualMachine: ligo.NewVM(),
	}

	// Register the Generator Plugin
	generator.PluginInit(worker.virtualMachine)
	defaultConfig(worker.virtualMachine)

	worker.spaces["overworld"] = function.NewSpace()
	worker.virtualMachine.Vars["space"] = ligo.Variable{
		Type: ligo.TypeStruct,
		Value: worker.spaces["overworld"],
	}


	dialer := func() Dialer {
		if config.User.Auth {
			return Dialer {
				TokenSource: auth.TokenSource,
			}
		} else {
			return Dialer {}
		}
	}()

	conn, err := dialer.Dial("raknet", config.Connection.RemoteAddress)
	if err != nil {
		pterm.Error.Println(err)
	}
	defer conn.Close()
	conn.worldConfig.operator = config.User.Operator
	conn.worldConfig.bot = config.User.Bot

	// Basic functions
	worker.virtualMachine.Funcs["plot"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		workSpace := vm.Vars["space"].Value.(*function.Space)
		if variable[0].Type == ligo.TypeArray {
			vec := variable[0].Value.([]function.Vector)
			var bar = progressbar.NewOptions(len(vec),
				progressbar.OptionSetWriter(BarWriter{
					conn: conn,
				}),
				progressbar.OptionEnableColorCodes(false),
				progressbar.OptionSetWidth(30),
				progressbar.OptionSetDescription("Building..."),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer: "-",
					SaucerHead: "+",
					BarStart: "[",
					BarEnd: "]",
					SaucerPadding: "=",
				}),
			)
			for _, v := range vec {
				conn.worldConfig.block.name = vm.Vars["block"].Value.(string)
				conn.worldConfig.block.data = vm.Vars["data"].Value.(int64)
				err := conn.SetBlock(function.AddVector(v, workSpace.GetPointer()))
				time.Sleep(time.Millisecond)
				if err != nil {
					return vm.Throw(fmt.Sprintf("setblock: Unable to setblock: %s", err))
				}
				bar.Add(1)
			}
		} else if variable[0].Type == ligo.TypeFloat {
			workSpace.Plot(variable[0].Value.(function.Vector))
		} else {
			return ligo.Variable{Type: ligo.TypeErr, Value: "plot function's first argument should be of a vector or vector slice type"}
		}
		return ligo.Variable {
			Type: ligo.TypeNil,
		}
	}

	worker.virtualMachine.Funcs["get"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		conn.SendCommand("gamerule sendcommandfeedback true", func(output *packet.CommandOutput) error {return nil})
		err := conn.SendCommand(fmt.Sprintf("execute %s ~ ~ ~ testforblock ~ ~ ~ air", conn.worldConfig.operator), func(output *packet.CommandOutput) error {
			pos, _ := function.SliceAtoi(output.OutputMessages[0].Parameters)
			if len(pos) != 3 {
				return errors.New("testforblock function have got wrong number of positions")
			} else {
				space := vm.Vars["space"].Value.(*function.Space)
				space.SetPointer(pos)
				_ = conn.Info(fmt.Sprintf("Position got: %v", pos))
				conn.SendCommand("gamerule sendcommandfeedback false", func(output *packet.CommandOutput) error {return nil})
			}
			return nil
		})
		if err != nil {
			return vm.Throw(fmt.Sprintf("SendCommand: %s", err))
		} else {
			return ligo.Variable{Type: ligo.TypeNil, Value: nil}
		}
	}

	worker.virtualMachine.Funcs["clear"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		text := ""
		for i := 0; i < 200 ; i++ {
			text += "\n"
		}
		conn.Info(text)
		return ligo.Variable{
			Type:  ligo.TypeNil,
			Value: nil,
		}
	}
	if err := conn.DoSpawn(); err == nil {
		pterm.Info.Println(fmt.Sprintf("Bot<%s> successfully spawned.", conn.worldConfig.bot))
		// Collector : Get Position
		eval, err := worker.virtualMachine.Eval(`(get)`)
		if err != nil {
			pterm.Error.Println(err)
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
				if conn.worldConfig.operator == p.SourceName {
					pterm.Info.Println(fmt.Sprintf("[%s] %s", p.SourceName, p.Message))
					value, err := worker.virtualMachine.Eval(p.Message)
					if err != nil {
						conn.Error(err.Error())
					} else {
						if err := conn.Info(fmt.Sprintf("==> %s", value.Value)) ; err != nil {
							pterm.Warning.Println(err)
						}
					}
				}
			}

		case *packet.CommandOutput:
			callback, ok := conn.callbacks[p.CommandOrigin.UUID.String()]
			// TODO : Handle !ok
			if ok {
				delete(conn.callbacks, p.CommandOrigin.UUID.String())
				if len(p.OutputMessages) > 0 {
					if !p.OutputMessages[0].Success {
						//pterm.Warning.Println(fmt.Sprintf("Unknown command: %s. Please check that the command exists and that you have permission to use it.", p.OutputMessages[0].Parameters))
					}
					err := callback(p)
					if err != nil {
						pterm.Warning.Println(err)
						// TODO : Handle error
					}
					continue
				}
			}
		case *packet.StructureTemplateDataResponse:
			data := p.StructureTemplate
			pterm.Info.Println(data)
		}

		// Write a packet to the connection: Similarly to ReadPacket, WritePacket will (only) return an error
		// if the connection is closed.
		p := &packet.RequestChunkRadius{ChunkRadius: 16}
		if err := conn.WritePacket(p); err != nil {
			break
		}

	}
}

func defaultConfig(vm *ligo.VM) {
	vm.Vars["block"] = ligo.Variable{
		Type:  ligo.TypeString,
		Value: "iron_block",
	}
	vm.Vars["data"] = ligo.Variable{
		Type:  ligo.TypeInt,
		Value: 0,
	}
}