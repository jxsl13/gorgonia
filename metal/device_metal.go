// ARM64 Metal GPU backend. SPEC §C8, T10. Build-tag gated.

//go:build metal && darwin && arm64

package metal

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Metal -framework Foundation -framework MetalPerformanceShaders -framework MetalPerformanceShadersGraph
#import <Metal/Metal.h>
#import <MetalPerformanceShaders/MetalPerformanceShaders.h>
#import <MetalPerformanceShadersGraph/MetalPerformanceShadersGraph.h>
#include <stdlib.h>
#include <string.h>

// Opaque handle bundling the device, command queue, and the elementwise
// compute pipelines. Held by Go as an uintptr (via a retained CFType-like ref).
typedef struct {
    id<MTLDevice>       dev;
    id<MTLCommandQueue> queue;
    id<MTLComputePipelineState> addPSO;
    id<MTLComputePipelineState> subPSO;
    id<MTLComputePipelineState> mulPSO;
} mtlCtx;

static const char* kSrc =
"#include <metal_stdlib>\n"
"using namespace metal;\n"
"kernel void vadd(device const float* a [[buffer(0)]], device const float* b [[buffer(1)]], device float* c [[buffer(2)]], uint i [[thread_position_in_grid]]) { c[i] = a[i] + b[i]; }\n"
"kernel void vsub(device const float* a [[buffer(0)]], device const float* b [[buffer(1)]], device float* c [[buffer(2)]], uint i [[thread_position_in_grid]]) { c[i] = a[i] - b[i]; }\n"
"kernel void vmul(device const float* a [[buffer(0)]], device const float* b [[buffer(1)]], device float* c [[buffer(2)]], uint i [[thread_position_in_grid]]) { c[i] = a[i] * b[i]; }\n";

static id<MTLComputePipelineState> makePSO(id<MTLDevice> dev, id<MTLLibrary> lib, const char* name) {
    id<MTLFunction> fn = [lib newFunctionWithName:[NSString stringWithUTF8String:name]];
    if (!fn) return nil;
    NSError* err = nil;
    return [dev newComputePipelineStateWithFunction:fn error:&err];
}

// mtlCtx contains Objective-C `id` types, which cgo cannot represent in Go.
// So the handle is passed to/from Go as an opaque void*.

// mtlNew creates the context. Returns NULL on failure.
static void* mtlNew() {
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return NULL;
    NSError* err = nil;
    id<MTLLibrary> lib = [dev newLibraryWithSource:[NSString stringWithUTF8String:kSrc] options:nil error:&err];
    if (!lib) return NULL;
    mtlCtx* c = (mtlCtx*)calloc(1, sizeof(mtlCtx));
    c->dev = dev;
    c->queue = [dev newCommandQueue];
    c->addPSO = makePSO(dev, lib, "vadd");
    c->subPSO = makePSO(dev, lib, "vsub");
    c->mulPSO = makePSO(dev, lib, "vmul");
    if (!c->queue || !c->addPSO || !c->subPSO || !c->mulPSO) { free(c); return NULL; }
    return (void*)c;
}

static void mtlFree(void* p) {
    if (!p) return;
    mtlCtx* c = (mtlCtx*)p;
    // ARC releases the held ObjC objects when the struct fields are cleared.
    c->dev = nil; c->queue = nil; c->addPSO = nil; c->subPSO = nil; c->mulPSO = nil;
    free(c);
}

static const char* mtlDeviceName(void* p) { return [[((mtlCtx*)p)->dev name] UTF8String]; }

// run dispatches one elementwise op (0=add,1=sub,2=mul) over n float32 elements,
// out <- a op b. Buffers are allocated shared (CPU/GPU) and freed each call.
static void mtlRun(void* p, int op, const float* a, const float* b, float* out, int n) {
    mtlCtx* c = (mtlCtx*)p;
    size_t bytes = (size_t)n * sizeof(float);
    id<MTLBuffer> ba = [c->dev newBufferWithBytes:a length:bytes options:MTLResourceStorageModeShared];
    id<MTLBuffer> bb = [c->dev newBufferWithBytes:b length:bytes options:MTLResourceStorageModeShared];
    id<MTLBuffer> bc = [c->dev newBufferWithLength:bytes options:MTLResourceStorageModeShared];

    id<MTLComputePipelineState> pso = c->addPSO;
    if (op == 1) pso = c->subPSO; else if (op == 2) pso = c->mulPSO;

    id<MTLCommandBuffer> cb = [c->queue commandBuffer];
    id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
    [enc setComputePipelineState:pso];
    [enc setBuffer:ba offset:0 atIndex:0];
    [enc setBuffer:bb offset:0 atIndex:1];
    [enc setBuffer:bc offset:0 atIndex:2];

    NSUInteger tg = pso.maxTotalThreadsPerThreadgroup;
    if (tg > (NSUInteger)n) tg = n;
    if (tg < 1) tg = 1;
    [enc dispatchThreads:MTLSizeMake(n, 1, 1) threadsPerThreadgroup:MTLSizeMake(tg, 1, 1)];
    [enc endEncoding];
    [cb commit];
    [cb waitUntilCompleted];

    memcpy(out, [bc contents], bytes);
    // ba, bb, bc released by ARC at scope exit.
}

// mtlMatMul computes C(MxN) = A(MxK) * B(KxN), row-major float32, on the GPU via
// MPSMatrixMultiplication.
static void mtlMatMul(void* p, const float* A, const float* B, float* C, int M, int N, int K) {
    mtlCtx* ctx = (mtlCtx*)p;
    size_t aBytes = (size_t)M*K*sizeof(float);
    size_t bBytes = (size_t)K*N*sizeof(float);
    size_t cBytes = (size_t)M*N*sizeof(float);

    id<MTLBuffer> ba = [ctx->dev newBufferWithBytes:A length:aBytes options:MTLResourceStorageModeShared];
    id<MTLBuffer> bb = [ctx->dev newBufferWithBytes:B length:bBytes options:MTLResourceStorageModeShared];
    id<MTLBuffer> bc = [ctx->dev newBufferWithLength:cBytes options:MTLResourceStorageModeShared];

    MPSMatrixDescriptor* da = [MPSMatrixDescriptor matrixDescriptorWithRows:M columns:K rowBytes:K*sizeof(float) dataType:MPSDataTypeFloat32];
    MPSMatrixDescriptor* db = [MPSMatrixDescriptor matrixDescriptorWithRows:K columns:N rowBytes:N*sizeof(float) dataType:MPSDataTypeFloat32];
    MPSMatrixDescriptor* dc = [MPSMatrixDescriptor matrixDescriptorWithRows:M columns:N rowBytes:N*sizeof(float) dataType:MPSDataTypeFloat32];

    MPSMatrix* ma = [[MPSMatrix alloc] initWithBuffer:ba descriptor:da];
    MPSMatrix* mb = [[MPSMatrix alloc] initWithBuffer:bb descriptor:db];
    MPSMatrix* mc = [[MPSMatrix alloc] initWithBuffer:bc descriptor:dc];

    MPSMatrixMultiplication* mm = [[MPSMatrixMultiplication alloc]
        initWithDevice:ctx->dev transposeLeft:NO transposeRight:NO
        resultRows:M resultColumns:N interiorColumns:K alpha:1.0 beta:0.0];

    id<MTLCommandBuffer> cb = [ctx->queue commandBuffer];
    [mm encodeToCommandBuffer:cb leftMatrix:ma rightMatrix:mb resultMatrix:mc];
    [cb commit];
    [cb waitUntilCompleted];

    memcpy(C, [bc contents], cBytes);
}

// mtlConv2D computes a 2D convolution on the GPU via MPSGraph.
// Layout: src NHWC [N,H,W,Cin], weights HWIO [KH,KW,Cin,Cout], out NHWC
// [N,OH,OW,Cout]. Explicit padding (pt,pb,pl,pr), stride (sx,sy). No bias.
// Returns 0 on success, nonzero on failure.
static int mtlConv2D(void* p,
        const float* src, const float* wts, float* out,
        int N, int H, int W, int Cin, int KH, int KW, int Cout,
        int sx, int sy, int pt, int pb, int pl, int pr,
        int OH, int OW) {
    mtlCtx* ctx = (mtlCtx*)p;

    size_t srcBytes = (size_t)N*H*W*Cin*sizeof(float);
    size_t wBytes   = (size_t)KH*KW*Cin*Cout*sizeof(float);
    size_t outBytes = (size_t)N*OH*OW*Cout*sizeof(float);

    id<MTLBuffer> bSrc = [ctx->dev newBufferWithBytes:src length:srcBytes options:MTLResourceStorageModeShared];
    id<MTLBuffer> bW   = [ctx->dev newBufferWithBytes:wts length:wBytes options:MTLResourceStorageModeShared];

    MPSGraph* g = [MPSGraph new];
    MPSGraphTensor* tSrc = [g placeholderWithShape:@[@(N),@(H),@(W),@(Cin)] dataType:MPSDataTypeFloat32 name:@"src"];
    MPSGraphTensor* tW   = [g placeholderWithShape:@[@(KH),@(KW),@(Cin),@(Cout)] dataType:MPSDataTypeFloat32 name:@"w"];

    MPSGraphConvolution2DOpDescriptor* desc =
        [MPSGraphConvolution2DOpDescriptor descriptorWithStrideInX:sx strideInY:sy
            dilationRateInX:1 dilationRateInY:1 groups:1
            paddingLeft:pl paddingRight:pr paddingTop:pt paddingBottom:pb
            paddingStyle:MPSGraphPaddingStyleExplicit
            dataLayout:MPSGraphTensorNamedDataLayoutNHWC
            weightsLayout:MPSGraphTensorNamedDataLayoutHWIO];
    if (!desc) return 1;

    MPSGraphTensor* tOut = [g convolution2DWithSourceTensor:tSrc weightsTensor:tW descriptor:desc name:nil];
    if (!tOut) return 2;

    MPSGraphTensorData* dSrc = [[MPSGraphTensorData alloc] initWithMTLBuffer:bSrc shape:@[@(N),@(H),@(W),@(Cin)] dataType:MPSDataTypeFloat32];
    MPSGraphTensorData* dW   = [[MPSGraphTensorData alloc] initWithMTLBuffer:bW shape:@[@(KH),@(KW),@(Cin),@(Cout)] dataType:MPSDataTypeFloat32];

    NSDictionary* results = [g runWithMTLCommandQueue:ctx->queue
        feeds:@{tSrc:dSrc, tW:dW}
        targetTensors:@[tOut] targetOperations:nil];

    MPSGraphTensorData* dOut = results[tOut];
    if (!dOut) return 3;
    [[dOut mpsndarray] readBytes:out strideBytes:nil];
    (void)outBytes;
    return 0;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// Device is a handle to the Metal GPU compute context (device + queue +
// pipelines). Not safe for concurrent use; create one per goroutine or guard it.
type Device struct {
	ctx unsafe.Pointer
}

// New creates a Metal compute device. Returns an error if no GPU / Metal is
// available or the kernels fail to compile.
func New() (*Device, error) {
	ctx := C.mtlNew()
	if ctx == nil {
		return nil, fmt.Errorf("metal: no GPU device or kernel compile failed")
	}
	d := &Device{ctx: ctx}
	runtime.SetFinalizer(d, (*Device).Close)
	return d, nil
}

// Name returns the GPU device name (e.g. "Apple M2 Pro").
func (d *Device) Name() string {
	return C.GoString(C.mtlDeviceName(d.ctx))
}

// Close releases GPU resources. Idempotent.
func (d *Device) Close() {
	if d.ctx != nil {
		C.mtlFree(d.ctx)
		d.ctx = nil
		runtime.SetFinalizer(d, nil)
	}
}

func (d *Device) run(op C.int, a, b []float32) ([]float32, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("metal: length mismatch %d != %d", len(a), len(b))
	}
	if d.ctx == nil {
		return nil, fmt.Errorf("metal: device closed")
	}
	n := len(a)
	out := make([]float32, n)
	if n == 0 {
		return out, nil
	}
	C.mtlRun(d.ctx, op,
		(*C.float)(unsafe.Pointer(&a[0])),
		(*C.float)(unsafe.Pointer(&b[0])),
		(*C.float)(unsafe.Pointer(&out[0])),
		C.int(n))
	return out, nil
}

// Add returns a + b elementwise, computed on the GPU.
func (d *Device) Add(a, b []float32) ([]float32, error) { return d.run(0, a, b) }

// Sub returns a - b elementwise, computed on the GPU.
func (d *Device) Sub(a, b []float32) ([]float32, error) { return d.run(1, a, b) }

// Mul returns a * b elementwise, computed on the GPU.
func (d *Device) Mul(a, b []float32) ([]float32, error) { return d.run(2, a, b) }

// MatMul computes C = A·B on the GPU (row-major float32) where A is M×K, B is
// K×N, and the returned C is M×N. Uses MPSMatrixMultiplication.
func (d *Device) MatMul(a, b []float32, m, n, k int) ([]float32, error) {
	if d.ctx == nil {
		return nil, fmt.Errorf("metal: device closed")
	}
	if len(a) != m*k {
		return nil, fmt.Errorf("metal: A length %d != M*K %d", len(a), m*k)
	}
	if len(b) != k*n {
		return nil, fmt.Errorf("metal: B length %d != K*N %d", len(b), k*n)
	}
	c := make([]float32, m*n)
	if m == 0 || n == 0 || k == 0 {
		return c, nil
	}
	C.mtlMatMul(d.ctx,
		(*C.float)(unsafe.Pointer(&a[0])),
		(*C.float)(unsafe.Pointer(&b[0])),
		(*C.float)(unsafe.Pointer(&c[0])),
		C.int(m), C.int(n), C.int(k))
	return c, nil
}

// Conv2D computes a 2D convolution on the GPU (MPSGraph). src is NHWC
// [n,h,w,cin], weights HWIO [kh,kw,cin,cout]. Returns NHWC [n,oh,ow,cout].
// Explicit padding (pt,pb,pl,pr) and stride (sx,sy); no bias.
func (d *Device) Conv2D(src, weights []float32, n, h, w, cin, kh, kw, cout, sx, sy, pt, pb, pl, pr int) (out []float32, oh, ow int, err error) {
	if d.ctx == nil {
		return nil, 0, 0, fmt.Errorf("metal: device closed")
	}
	if len(src) != n*h*w*cin {
		return nil, 0, 0, fmt.Errorf("metal: src len %d != n*h*w*cin %d", len(src), n*h*w*cin)
	}
	if len(weights) != kh*kw*cin*cout {
		return nil, 0, 0, fmt.Errorf("metal: weights len %d != kh*kw*cin*cout %d", len(weights), kh*kw*cin*cout)
	}
	oh = (h+pt+pb-kh)/sy + 1
	ow = (w+pl+pr-kw)/sx + 1
	if oh <= 0 || ow <= 0 {
		return nil, 0, 0, fmt.Errorf("metal: non-positive output dims %dx%d", oh, ow)
	}
	out = make([]float32, n*oh*ow*cout)
	rc := C.mtlConv2D(d.ctx,
		(*C.float)(unsafe.Pointer(&src[0])),
		(*C.float)(unsafe.Pointer(&weights[0])),
		(*C.float)(unsafe.Pointer(&out[0])),
		C.int(n), C.int(h), C.int(w), C.int(cin), C.int(kh), C.int(kw), C.int(cout),
		C.int(sx), C.int(sy), C.int(pt), C.int(pb), C.int(pl), C.int(pr),
		C.int(oh), C.int(ow))
	if rc != 0 {
		return nil, 0, 0, fmt.Errorf("metal: conv2d failed (code %d)", int(rc))
	}
	return out, oh, ow, nil
}
