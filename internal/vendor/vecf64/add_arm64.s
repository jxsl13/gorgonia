// MOD(jxsl13): ARM64 NEON elementwise float64 add/sub/mul. T20, SPEC §V11.
// 2 lanes (V.D2 = 128-bit) per iteration + scalar remainder. Bit-exact vs the
// scalar reference (same IEEE-754 round-to-nearest-even).
//
// Go's arm64 assembler lacks vector FP mnemonics, so FADD/FSUB/FMUL (vector,
// 2D, V0=V0 op V1) are WORD-encoded:
//   FADD V0.2D,V0.2D,V1.2D = 0x4E61D400
//   FSUB V0.2D,V0.2D,V1.2D = 0x4EE1D400
//   FMUL V0.2D,V0.2D,V1.2D = 0x6E61DC00
// Encodings runtime-verified ([1,2] op [10,20]).

//go:build arm64
// +build arm64

#include "textflag.h"

// func addAsm(a, b []float64)
TEXT ·addAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0
	MOVD a_len+8(FP), R2
	MOVD b_base+24(FP), R1

	LSR  $1, R2, R3         // R3 = len/2 (2-wide blocks)
	CBZ  R3, add_rem
add_loop:
	VLD1 (R0), [V0.D2]
	VLD1 (R1), [V1.D2]
	WORD $0x4E61D400       // FADD V0.2D,V0.2D,V1.2D
	VST1 [V0.D2], (R0)
	ADD  $16, R0
	ADD  $16, R1
	SUB  $1, R3
	CBNZ R3, add_loop
add_rem:
	AND  $1, R2, R4        // R4 = len%2
	CBZ  R4, add_done
	FMOVD (R0), F0
	FMOVD (R1), F1
	FADDD F1, F0, F0
	FMOVD F0, (R0)
add_done:
	RET

// func subAsm(a, b []float64)
TEXT ·subAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0
	MOVD a_len+8(FP), R2
	MOVD b_base+24(FP), R1

	LSR  $1, R2, R3
	CBZ  R3, sub_rem
sub_loop:
	VLD1 (R0), [V0.D2]
	VLD1 (R1), [V1.D2]
	WORD $0x4EE1D400       // FSUB V0.2D,V0.2D,V1.2D
	VST1 [V0.D2], (R0)
	ADD  $16, R0
	ADD  $16, R1
	SUB  $1, R3
	CBNZ R3, sub_loop
sub_rem:
	AND  $1, R2, R4
	CBZ  R4, sub_done
	FMOVD (R0), F0
	FMOVD (R1), F1
	FSUBD F1, F0, F0
	FMOVD F0, (R0)
sub_done:
	RET

// func mulAsm(a, b []float64)
TEXT ·mulAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0
	MOVD a_len+8(FP), R2
	MOVD b_base+24(FP), R1

	LSR  $1, R2, R3
	CBZ  R3, mul_rem
mul_loop:
	VLD1 (R0), [V0.D2]
	VLD1 (R1), [V1.D2]
	WORD $0x6E61DC00       // FMUL V0.2D,V0.2D,V1.2D
	VST1 [V0.D2], (R0)
	ADD  $16, R0
	ADD  $16, R1
	SUB  $1, R3
	CBNZ R3, mul_loop
mul_rem:
	AND  $1, R2, R4
	CBZ  R4, mul_done
	FMOVD (R0), F0
	FMOVD (R1), F1
	FMULD F1, F0, F0
	FMOVD F0, (R0)
mul_done:
	RET
