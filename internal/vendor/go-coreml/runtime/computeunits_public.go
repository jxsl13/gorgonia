// MOD(jxsl13): upstream keeps the ComputeUnits constants in internal/bridge, so
// external callers cannot pass them to WithComputeUnits (the README example does
// not compile outside this module). Re-export them publicly so our coreml/
// wrapper can select compute units. SPEC §C12, T15.

package runtime

import "github.com/gomlx/go-coreml/internal/bridge"

// ComputeUnits selects which Apple compute units CoreML may use.
type ComputeUnits = bridge.ComputeUnits

// Public aliases for the internal bridge constants.
const (
	ComputeAll       = bridge.ComputeAll       // ANE + GPU + CPU (default)
	ComputeCPUOnly   = bridge.ComputeCPUOnly   // CPU only (debugging)
	ComputeCPUAndGPU = bridge.ComputeCPUAndGPU // CPU + GPU (skip ANE)
	ComputeCPUAndANE = bridge.ComputeCPUAndANE // CPU + ANE (skip GPU)
)
