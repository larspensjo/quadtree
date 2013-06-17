package quadtree

import (
	"math"
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

// Measure time to add objects.
func BenchmarkAdd(t *testing.B) {
	t.StopTimer()
	tree := basicTree()
	list := make([]o, t.N)
	for _, obj := range list {
		obj[0] = rand.Float64()
		obj[1] = rand.Float64()
	}
	t.StartTimer()
	for _, obj := range list {
		tree.Add(&obj)
	}
}

// Measure time to move objects.
func BenchmarkMove(t *testing.B) {
	t.StopTimer()
	tree := basicTree()
	list := make([]o, t.N)
	for i := range list {
		list[i][0] = rand.Float64()
		list[i][1] = rand.Float64()
	}
	for i := range list {
		tree.Add(&list[i])
	}
	delta := 1 / math.Sqrt(float64(t.N)) // Distance to move
	t.StartTimer()
	for i := range list {
		obj := &list[i]
		newPos := obj.GetCurrentPosition()
		newPos[0] += (rand.Float64() - 0.5) * delta
		newPos[1] += (rand.Float64() - 0.5) * delta
		tree.Move(obj, newPos)
	}
}

// Find all objects near another object
func BenchmarkFind(t *testing.B) {
	t.StopTimer()
	tree := basicTree()
	list := make([]o, t.N)
	for i := range list {
		list[i][0] = rand.Float64()
		list[i][1] = rand.Float64()
	}
	for i := range list {
		tree.Add(&list[i])
	}
	delta := 2 / math.Sqrt(float64(t.N)) // Distance to search
	t.StartTimer()
	tot := 0
	for i := range list {
		obj := &list[i]
		result := tree.FindNearObjects(obj.GetCurrentPosition(), delta)
		tot += len(result)
	}
	// t.Log(t.N, "objects: found", float64(tot)/float64(t.N), "on average")
}
