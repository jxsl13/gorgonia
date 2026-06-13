module github.com/jxsl13/gorgonia

go 1.26

toolchain go1.26.4

require (
	github.com/awalterschulze/gographviz v2.0.3+incompatible
	github.com/chewxy/hm v1.0.0
	github.com/chewxy/math32 v1.11.2
	github.com/go-gota/gota v0.12.0
	github.com/gomlx/go-coreml v0.0.0-20260301010621-8fdf6ad8655e
	github.com/leesper/go_rng v0.0.0-20190531154944-a612b043e353
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.11.1
	github.com/xtgo/set v1.0.0
	gonum.org/v1/gonum v0.17.0
	gonum.org/v1/netlib v0.0.0-20230729102104-8b8060e7531f
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gorgonia.org/cu v0.9.6
	gorgonia.org/dawson v1.2.0
	gorgonia.org/tensor v0.9.24
	gorgonia.org/vecf32 v0.9.0
	gorgonia.org/vecf64 v0.9.0
)

require (
	github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20231121144256-b99613f794b6 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorgonia.org/gorgonia v0.9.18 // indirect
)

replace gorgonia.org/vecf32 => ./internal/vendor/vecf32

replace gorgonia.org/vecf64 => ./internal/vendor/vecf64

replace github.com/gomlx/go-coreml => ./internal/vendor/go-coreml
