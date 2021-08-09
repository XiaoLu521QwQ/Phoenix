package generator

import "phoenix/minecraft/ligo"

func PluginInit(vm *ligo.VM){
	vm.Funcs["circle"] = Circle

}

// Circle : (circle radius inner radius height facing)
func Circle(vm *ligo.VM, a ...ligo.Variable) ligo.Variable {
	pos = vm.Vars["position"]
}
