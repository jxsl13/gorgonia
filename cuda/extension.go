//go:build cuda
// +build cuda

package cuda

import (
	"fmt"

	"gorgonia.org/cu"
)

// this file relates to code that allows you to extend Engine

// LoadCUDAFunc loads a string representing a CUDA PTX file into the engine, giving it the universe of computing functions.
func (e *Engine) LoadCUDAFunc(moduleName, data string, funcs []string) (err error) {
	fns := e.f
	if fns == nil {
		fns = make(map[string]cu.Function)
	}
	if err = cu.SetCurrentContext(e.c.Context.CUDAContext()); err != nil {
		return fmt.Errorf("Unable to set current context when loading module %q at device %v: %w", moduleName, e.d, err)
	}

	var mod cu.Module
	if mod, err = cu.LoadData(data); err != nil {
		return fmt.Errorf("Failed to load module %q data for Device %v context %x: %w", moduleName, e.d, e.c, err)
	}

	for _, name := range funcs {
		var fn cu.Function
		if fn, err = mod.Function(name); err != nil {
			return fmt.Errorf("Unable to get function %q in Device %v context %x: %w", name, e.d, e.c, err)
		}
		fqn := fmt.Sprintf("%v.%v", moduleName, name)
		fns[fqn] = fn
	}
	if e.m == nil {
		e.m = make(map[string]cu.Module)
	}
	e.m[moduleName] = mod
	e.f = fns
	return nil
}
