//go:build !debug

package gorgonia

// DEBUG indicates if this build is in debug mode. It is not.
const DEBUG = false

const (
	compileDev        = false
	shapeInferenceDev = false
	typeSystemDev     = false
	symdiffDev        = false
	autodiffDev       = false
	machineDev        = false
	stabilizationDev  = false
	cudaDev           = false
	allocatorDev      = false
)

func tabcount() int { return 0 }

func enterLogScope()                           {}
func leaveLogScope()                           {}
func logf(format string, others ...any)        {}
func compileLogf(format string, attrs ...any)  {}
func shapeLogf(format string, attrs ...any)    {}
func typeSysLogf(format string, attrs ...any)  {}
func symdiffLogf(format string, attrs ...any)  {}
func autodiffLogf(format string, attrs ...any) {}
func machineLogf(format string, attrs ...any)  {}
func stabLogf(format string, attrs ...any)     {}
func solverLogf(format string, attrs ...any)   {}
func cudaLogf(format string, attrs ...any)     {}
func allocatorLogf(format string, attr ...any) {}
func recoverFrom(format string, attrs ...any)  {}

// GraphCollisionStats returns the collisions in the graph only when built with the debug tag, otherwise it's a noop that returns 0
func GraphCollisionStats() (int, int, int) { return 0, 0, 0 }

func incrCC() {}
func incrEC() {}
func incrNN() {}

/* Compilation related debug utility functions/methods*/
func logCompileState(name string, g *ExprGraph, df *dataflow) {}

/* Analysis Debug Utility Functions/Methods */
func (df *dataflow) debugIntervals(sorted Nodes) {}
