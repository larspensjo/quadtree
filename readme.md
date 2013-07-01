Quadtree
========
Maintain a list of objects.
The purpose is to keep the cost down for finding near objects.
See [Wikipedia Quadtree](http://en.wikipedia.org/wiki/Quadtree).

* All objects must embed the type 'quadtree.Handle'.
* The package is thread safe.
* The quadtree will automatically grow in size, as needed.

For library details, see [Quadtree documentation](http://godoc.org/github.com/larspensjo/quadtree).

An example taken from the automatic testing:
```
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
```
