Quadtree
========
Maintain a list of objects.
The purpose is to keep the cost down for finding near objects.
See [Wikipedia Quadtree](http://en.wikipedia.org/wiki/Quadtree).

* All objects must embed the type 'QuadtreePosition'.
* The package is thread safe as long as the position of objects in the quadtree does not change other than when SetPosition() callback is called.

For library details, see [Quadtree documentation](http://godoc.org/github.com/larspensjo/quadtree).
