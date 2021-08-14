package minecraft

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"phoenix/minecraft/auth"
	"phoenix/minecraft/function"
	"phoenix/minecraft/function/generator"
	"phoenix/minecraft/function/std"
	"phoenix/minecraft/ligo"
	"phoenix/minecraft/protocol/packet"
	"time"
)


type Worker struct {
	spaces         map[string]*function.Space
	VirtualMachine *ligo.VM
}

func (w *Worker) GetSpace(name string) *function.Space {
	return w.spaces[name]
}

func Run(path string) {
	config := function.ReadConfig(path)
	if config.Debug.Enabled {
		pterm.EnableDebugMessages()
	}
	
	// Init Connection :: Start 
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
	// Init Connection :: End 

	if config.Lib.Std {
		std.StdInit(conn.  Worker.VirtualMachine)
	}
	// Register the Generator Plugin
	generator.PluginInit(conn.Worker.VirtualMachine)
	if err := LoadScript(conn.Worker.VirtualMachine, config.Lib.Script) ; err != nil {
		pterm.Error.Println(err)
	}
	defaultConfig(conn.Worker.VirtualMachine)
	
	conn.WorldConfig.Operator = config.User.Operator
	conn.WorldConfig.bot = config.User.Bot

	// Basic Functions Init
	InitWorker(conn)


	if err := conn.DoSpawn(); err == nil {
		pterm.Info.Println(fmt.Sprintf("Bot<%s> successfully spawned.", conn.identityData.DisplayName))
		// Collector : Get Position
		eval, err := conn.Worker.VirtualMachine.Eval(`(get)`)
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
				if conn.WorldConfig.Operator == p.SourceName {
					pterm.Info.Println(fmt.Sprintf("[%s] %s", p.SourceName, p.Message))
					value, err := conn.Worker.VirtualMachine.Eval(p.Message)
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
	var data int64
	data = 0
	vm.Vars["data"] = ligo.Variable{
		Type:  ligo.TypeInt,
		Value: data,
	}
}

func InitWorker(conn *Conn) {
	conn.Worker.VirtualMachine.Funcs["get"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
		conn.SendCommand("gamerule sendcommandfeedback true", func(output *packet.CommandOutput) error {return nil})
		err := conn.SendCommand(fmt.Sprintf("execute %s ~ ~ ~ testforblock ~ ~ ~ air", conn.WorldConfig.Operator), func(output *packet.CommandOutput) error {
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

	conn.Worker.VirtualMachine.Funcs["plot"] = func(vm *ligo.VM, variable ...ligo.Variable) ligo.Variable {
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
				conn.WorldConfig.block.name = vm.Vars["block"].Value.(string)
				conn.WorldConfig.block.data = vm.Vars["data"].Value.(int64)
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
}

func LoadScript(vm *ligo.VM, paths []string) error {
	for n, path := range paths {
		fileName := filepath.Base(paths[n])
		fileName = fileName[:len(fileName) - len(filepath.Ext(fileName))]
		if content, err := os.Open(path) ; err != nil {
			return err
		} else if err := vm.LoadReader(content) ; err != nil {
			return errors.New(fmt.Sprintf("Error loading script [%s]: %s", fileName, err))
		}
		pterm.Info.Println(fmt.Sprintf("Successfully loaded Script [%s] ",fileName))
	}
	return nil
}