package quadtree

import (
	"testing"
)

func TestInitial(t *testing.T) {
	lowerLeft := [2]float64{0, 0}
	upperRight := [2]float64{1, 1}
	tree := MakeQuadtree(lowerLeft, upperRight, 6)
	if !tree.Empty() || tree.hasChildren || tree.numObjects > 0 || len(tree.objects) != 0 {
		t.Error("Initial tree not empty")
	}
}
