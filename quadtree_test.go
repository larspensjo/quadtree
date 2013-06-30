package quadtree

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

type o struct {
	Handle
}

func basicTree() *Quadtree {
	lowerLeft := Twof{0, 0}
	upperRight := Twof{1, 1}
	return MakeQuadtree(lowerLeft, upperRight)
}

func TestInitial(t *testing.T) {
	tree := basicTree()
	if !tree.Empty() || tree.hasChildren || tree.quadtree.numObjects > 0 || len(tree.objects) != 0 {
		t.Error("Initial tree not empty")
	}

	var x1 o
	tree.Add(&x1, Twof{1.0, 2.0})
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
	positions := make([]Twof, t.N)
	for _, obj := range positions {
		obj[0] = rand.Float64()
		obj[1] = rand.Float64()
	}
	t.StartTimer()
	for i := range list {
		tree.Add(&list[i], positions[i])
	}
}

// Measure time to move objects.
func BenchmarkMove(t *testing.B) {
	t.StopTimer()
	tree := basicTree()
	list := make([]o, t.N)
	positions := make([]Twof, t.N)
	for i := range list {
		positions[i][0] = rand.Float64()
		positions[i][1] = rand.Float64()
	}
	for i := range list {
		tree.Add(&list[i], positions[i])
	}
	delta := 1 / math.Sqrt(float64(t.N)) // Distance to move
	t.StartTimer()
	for i := range list {
		obj := &list[i]
		newPos := obj.GetQuadtreePosition()
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
	positions := make([]Twof, t.N)
	for i := range list {
		positions[i][0] = rand.Float64()
		positions[i][1] = rand.Float64()
	}
	for i := range list {
		tree.Add(&list[i], positions[i])
	}
	delta := 2 / math.Sqrt(float64(t.N)) // Distance to search
	t.StartTimer()
	tot := 0
	for i := range list {
		obj := &list[i]
		result := tree.FindNearObjects(obj.GetQuadtreePosition(), delta)
		tot += len(result)
	}
	// t.Log(t.N, "objects: found", float64(tot)/float64(t.N), "on average")
}

type ball struct {
	Handle
	// Add other attributes here
}

func ExampleBalls() {
	upperLeft := Twof{0, 0}
	lowerRight := Twof{1, 1}
	tree := MakeQuadtree(upperLeft, lowerRight)
	// Create 10 balls and add them to the quadtree
	for i := 0; i < 10; i++ {
		var b ball
		tree.Add(&b, Twof{float64(i) / 10.0, 0})
	}
	list := tree.FindNearObjects(Twof{0.5, 0.1}, 0.2)
	fmt.Println("Found", len(list))
	// Output: Found 3
}
