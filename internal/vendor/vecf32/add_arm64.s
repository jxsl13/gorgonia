// MOD(jxsl13): ARM64 NEON elementwise float32 add/sub/mul. T20, SPEC §V11.
// 4 lanes (V.S4 = 128-bit) per iteration + scalar remainder. Results are
// bit-exact vs the scalar reference (same IEEE-754 round-to-nearest-even).
//
// Go's arm64 assembler lacks vector FP mnemonics (VFADD/VFSUB/VFMUL), so the
// FADD/FSUB/FMUL (vector, 4S, V0=V0 op V1) ops are WORD-encoded:
//   FADD V0.4S,V0.4S,V1.4S = 0x4E21D400
//   FSUB V0.4S,V0.4S,V1.4S = 0x4EA1D400
//   FMUL V0.4S,V0.4S,V1.4S = 0x6E21DC00
// Encodings runtime-verified ([1,2,3,4] op [10,20,30,40]).

//go:build arm64
// +build arm64

#include "textflag.h"

// func addAsm(a, b []float32)
TEXT ·addAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0   // &a[0]
	MOVD a_len+8(FP), R2    // len(a)
	MOVD b_base+24(FP), R1  // &b[0]

	LSR  $2, R2, R3         // R3 = len/4  (number of 4-wide blocks)
	CBZ  R3, add_rem
add_loop:
	VLD1  (R0), [V0.S4]
	VLD1  (R1), [V1.S4]
	WORD  $0x4E21D400  // FADD V0.4S,V0.4S,V1.4S
	VST1  [V0.S4], (R0)
	ADD   $16, R0
	ADD   $16, R1
	SUB   $1, R3
	CBNZ  R3, add_loop
add_rem:
	AND  $3, R2, R4         // R4 = len%4
	CBZ  R4, add_done
add_remloop:
	FMOVS (R0), F0
	FMOVS (R1), F1
	FADDS F1, F0, F0
	FMOVS F0, (R0)
	ADD   $4, R0
	ADD   $4, R1
	SUB   $1, R4
	CBNZ  R4, add_remloop
add_done:
	RET

// func subAsm(a, b []float32)
TEXT ·subAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0
	MOVD a_len+8(FP), R2
	MOVD b_base+24(FP), R1

	LSR  $2, R2, R3
	CBZ  R3, sub_rem
sub_loop:
	VLD1  (R0), [V0.S4]
	VLD1  (R1), [V1.S4]
	WORD  $0x4EA1D400  // FSUB V0.4S,V0.4S,V1.4S
	VST1  [V0.S4], (R0)
	ADD   $16, R0
	ADD   $16, R1
	SUB   $1, R3
	CBNZ  R3, sub_loop
sub_rem:
	AND  $3, R2, R4
	CBZ  R4, sub_done
sub_remloop:
	FMOVS (R0), F0
	FMOVS (R1), F1
	FSUBS F1, F0, F0
	FMOVS F0, (R0)
	ADD   $4, R0
	ADD   $4, R1
	SUB   $1, R4
	CBNZ  R4, sub_remloop
sub_done:
	RET

// func mulAsm(a, b []float32)
TEXT ·mulAsm(SB), NOSPLIT, $0-48
	MOVD a_base+0(FP), R0
	MOVD a_len+8(FP), R2
	MOVD b_base+24(FP), R1

	LSR  $2, R2, R3
	CBZ  R3, mul_rem
mul_loop:
	VLD1  (R0), [V0.S4]
	VLD1  (R1), [V1.S4]
	WORD  $0x6E21DC00  // FMUL V0.4S,V0.4S,V1.4S
	VST1  [V0.S4], (R0)
	ADD   $16, R0
	ADD   $16, R1
	SUB   $1, R3
	CBNZ  R3, mul_loop
mul_rem:
	AND  $3, R2, R4
	CBZ  R4, mul_done
mul_remloop:
	FMOVS (R0), F0
	FMOVS (R1), F1
	FMULS F1, F0, F0
	FMOVS F0, (R0)
	ADD   $4, R0
	ADD   $4, R1
	SUB   $1, R4
	CBNZ  R4, mul_remloop
mul_done:
	RET
