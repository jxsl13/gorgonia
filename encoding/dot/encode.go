package dot

import (
	gonumDot "gonum.org/v1/gonum/graph/encoding/dot"
	"github.com/jxsl13/gorgonia"
)

// Marshal the graph in a dot (graphviz)
// This methods also generates the subgraphs
func Marshal(g *gorgonia.ExprGraph) ([]byte, error) {
	dg, err := generateDotGraph(g)
	if err != nil {
		return nil, err
	}

	return gonumDot.Marshal(dg, "", "", "\t")
}
