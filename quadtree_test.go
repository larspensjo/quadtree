package quadtree

import (
	"math/rand"
	"testing"
)

type o [2]float64

func (obj *o) GetCurrentPosition() [2]float64 {
	return *obj
}

func (obj *o) SetPosition(newPos [2]float64) {
	*obj = newPos
}

func basicTree() *Quadtree {
	lowerLeft := [2]float64{0, 0}
	upperRight := [2]float64{1, 1}
	return MakeQuadtree(lowerLeft, upperRight)
}

func TestInitial(t *testing.T) {
	tree := basicTree()
	if !tree.Empty() || tree.hasChildren || tree.numObjects > 0 || len(tree.objects) != 0 {
		t.Error("Initial tree not empty")
	}

	x1 := o{1.0, 2.0}
	tree.Add(&x1)
	if tree.numObjects != 1 || tree.Empty() {
		t.Error("Expected size 1")
	}

	tree.Remove(&x1)
	if !tree.Empty() || tree.hasChildren || tree.numObjects > 0 || len(tree.objects) != 0 {
		t.Error("Tree should be empty")
	}
}

func BenchmarkAdd(t *testing.B) {
	tree := basicTree()
	list := make([]o, t.N)
	for _, obj := range list {
		obj[0] = rand.Float64()
		obj[1] = rand.Float64()
	}
	for _, obj := range list {
		tree.Add(&obj)
	}
}

func BenchmarkMove(t *testing.B) {
	tree := basicTree()
	list := make([]o, t.N)
	for i := range list {
		list[i][0] = rand.Float64()
		list[i][1] = rand.Float64()
	}
	for i := range list {
		tree.Add(&list[i])
	}
	const delta = 0.001
	for iter := 0; iter < 100; iter++ {
		// log.Println("Iter", iter)
		for i := range list {
			obj := &list[i]
			newPos := obj.GetCurrentPosition()
			newPos[0] += (rand.Float64() - 0.5) * delta
			newPos[1] += (rand.Float64() - 0.5) * delta
			tree.Move(obj, newPos)
		}
	}
	t.Log(tree)
}
