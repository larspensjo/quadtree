// Copyright 2012-1013 Lars Pensjö
//
// Ephenation is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, version 3.
//
// Ephenation is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// See <http://www.gnu.org/licenses/>.
//

// Package quadtree is used for keeping track of what objects are close to each other.
// The cost of checking all possible objects would grow with the square of the number of
// objects, so this package will recursively divide a volume into 2 (in every dimension, giving 4
// sub cubes) when the number of objects exceeds a certain limit.
package quadtree

import (
	"log"
	"sync"
)

// Depth estimate:
// 10000 players
// Divided into 4^6 squares gives 2 players per square.
// Allow for higher concentration at some places, and it should still be enough.

const (
	maxQuadtreeDepth      = 8   // Do not make more levels below this
	minObjectsPerQuadtree = 5   // Lower limit of the number of objects before collapsing
	maxObjectsPerQuadtree = 10  // Upper limit of the number of objects before expanding
	expandFactor          = 1.3 // How much the are is expanded when the volume is too small
)

// Twof is the two dimensional type used as position.
type Twof [2]float64

// An object that can be stored in a Quadtree must embed this type anonymously.
type Handle struct {
	pos Twof // This is private to Quadtree
}

// Compute the squared distance between two points
func computeDist2(from, to Twof) float64 {
	dx := from[0] - to[0]
	dy := from[1] - to[1]
	return dx*dx + dy*dy
}

// The objects that are managed shall fulfill this interface
type Object interface {
	getCurrentPosition() Twof // The current coordinate
	setPosition(Twof)         // Callback that requests the position to be updated
}

func (p *Handle) getCurrentPosition() Twof {
	return p.pos
}

func (p *Handle) setPosition(n Twof) {
	p.pos = n
}

type quadtree struct {
	hasChildren bool            // Flag indicating whether the sub tree is used
	depth       int             // Depth of this node
	numObjects  int             // Sum of all objects in all children
	corner1     Twof            // Lower left corner
	corner2     Twof            // upper right corner
	center      Twof            // middle of the square
	children    [2][2]*quadtree // Sub tree used if enough number of objects
	objects     []Object        // List of objects used if sub tree is not used
}

// Use MakeQuadtree() to get one.
type Quadtree struct {
	quadtree
	mutex sync.RWMutex
}

// Check if the Quadtree is big enough to contain the given position. This is done by simply making
// a bigger initial square and moving all objects to the new one. Not a very cheap solution, but
// it is expected to be done rarely.
func (t *Quadtree) checkExpand(tf Twof) {
	changed := false
	newCorner1 := t.corner1
	newCorner2 := t.corner2
	for i := 0; i < 2; i++ {
		if tf[i] < t.corner1[i] {
			changed = true
			newCorner1[i] = t.corner2[i] - (t.corner2[i]-tf[i])*expandFactor
		}
		if tf[i] > t.corner2[i] {
			changed = true
			newCorner2[i] = t.corner1[i] + (tf[i]-t.corner1[i])*expandFactor
		}
	}
	if !changed {
		return
	}
	t.destroyChildren() // This will move all objects to the root.
	t.corner1 = newCorner1
	t.corner2 = newCorner2
	// Next time an object is added, the tree will expand again.
}

// Empty returns true if this Quadtree is empty. Used for debugging and testing.
func (t *Quadtree) Empty() bool {
	// No need to lock for this operation.
	return t.numObjects == 0 && !t.hasChildren && len(t.objects) == 0
}

// Initialize a quadtree
func (t *quadtree) init(c1, c2 Twof, depth int) {
	t.corner1 = c1
	t.corner2 = c2
	t.center = Twof{(c1[0] + c2[0]) / 2, (c1[1] + c2[1]) / 2}
	t.depth = depth
}

// MakeQuadtree creates a Quadtree.
// 'c1' is the corner with the smaller values,
// 'c2' is the corner with the bigger values.
func MakeQuadtree(c1, c2 Twof) *Quadtree {
	var t Quadtree
	t.init(c1, c2, 0)
	return &t
}

// Local version, making a sub node
func makequadtree(c1, c2 Twof, depth int) *quadtree {
	var t quadtree
	t.init(c1, c2, depth)
	return &t
}

// Adds or removes an object from the children. The size of objects are considered to be 0,
// which means an object can only be located in one child.
func (t *quadtree) fileObject(o Object, add bool) {
	// Figure out in what child the object belongs
	c := o.getCurrentPosition()
	for x := 0; x < 2; x++ {
		if x == 0 {
			if c[0] > t.center[0] {
				continue
			}
		} else if c[0] < t.center[0] {
			continue
		}

		for y := 0; y < 2; y++ {
			if y == 0 {
				if c[1] > t.center[1] {
					continue
				}
			} else if c[1] < t.center[1] {
				continue
			}

			// Add or remove the object
			if add {
				t.children[x][y].add(o)
			} else {
				t.children[x][y].remove(o)
			}
			return
		}
	}
}

// Take a leaf in the quadtree, add children, and move all objects to the children.
func (t *quadtree) makeChildren() {
	for x := 0; x < 2; x++ {
		var minX, maxX float64
		if x == 0 {
			minX = t.corner1[0]
			maxX = t.center[0]
		} else {
			minX = t.center[0]
			maxX = t.corner2[0]
		}

		for y := 0; y < 2; y++ {
			var minY, maxY float64
			if y == 0 {
				minY = t.corner1[1]
				maxY = t.center[1]
			} else {
				minY = t.center[1]
				maxY = t.corner2[1]
			}

			t.children[x][y] = makequadtree(Twof{minX, minY}, Twof{maxX, maxY}, t.depth+1)
		}
	}

	// Add all objects to the new children and remove them from "objects"
	for _, it := range t.objects {
		t.fileObject(it, true) // Use previous pos as the object may be moving asynchronously
	}
	t.objects = nil
	t.hasChildren = true
}

// Destroys the children of this, and moves all objects in its descendants
// to the "objects" set
func (t *quadtree) destroyChildren() {
	// Move all objects in descendants of this to the "objects" set
	t.objects = make([]Object, 0, t.numObjects)
	t.collectObjects(&t.objects)

	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			t.children[x][y] = nil
		}
	}

	t.hasChildren = false
}

// Removes the specified object
func (t *quadtree) remove(o Object) {
	t.numObjects--
	if t.numObjects < 0 {
		log.Panicln(">>>>Quadtree:remove numObjects < 0", t)
	}

	if t.hasChildren && t.numObjects < minObjectsPerQuadtree {
		t.destroyChildren()
	}

	if t.hasChildren {
		t.fileObject(o, false)
	} else {
		// Find o in the local list
		for i, o2 := range t.objects {
			if o2 == o {
				// Found it
				if last := len(t.objects) - 1; i == last {
					t.objects = t.objects[:last]
				} else {
					// Move the last element to this position
					t.objects[i] = t.objects[last]
					t.objects = t.objects[:last]
				}
				return
			}
		}
		log.Panicln("Quadtree:remove failed to find object")
	}
}

// Removes the specified object 'o'.
func (t *Quadtree) Remove(o Object) {
	t.mutex.Lock()
	t.remove(o)
	t.mutex.Unlock()
}

// Add an object 'o' at position 'c'.
func (t *Quadtree) Add(o Object, c Twof) {
	t.mutex.Lock()
	o.setPosition(c)
	t.checkExpand(c)
	t.add(o)
	t.mutex.Unlock()
}

// Add an object
func (t *quadtree) add(o Object) {
	t.numObjects++
	if !t.hasChildren && t.depth < maxQuadtreeDepth && t.numObjects > maxObjectsPerQuadtree {
		t.makeChildren()
	}

	if t.hasChildren {
		t.fileObject(o, true) // Use previous pos as the object may be moving asynchronously
	} else {
		t.objects = append(t.objects, o)
	}
}

// Test that an object, at the specified position, is already in the quadtree where it should be.
func (t *quadtree) testPresent(o Object, pos Twof) bool {
	if !t.hasChildren {
		// There are no children to this tree, which means the object should be in the list of objects.
		for _, o2 := range t.objects {
			if o2 == o {
				// Found it
				return true
			}
		}
		return false
	}
	// Figure out in which child(ren) the object belongs
	for x := 0; x < 2; x++ {
		if x == 0 {
			if pos[0] > t.center[0] {
				continue
			}
		} else if pos[0] < t.center[0] {
			continue
		}

		for y := 0; y < 2; y++ {
			if y == 0 {
				if pos[1] > t.center[1] {
					continue
				}
			} else if pos[1] < t.center[1] {
				continue
			}

			return t.children[x][y].testPresent(o, pos)
		}
	}
	// The object has to be somewhere in the quadtree. This shall never happen!
	log.Panicln("Quadtree.testPresent failed", o, pos, t)
	return false
}

// Move the position of an object 'o' to position 'to'.
func (t *Quadtree) Move(o Object, to Twof) {
	// Assume the obect was moved to another part of the quadtree
	treeChanged := true
	// Usually, the object will not be moved from one part of the quadtree to another. Do a test if that is
	// the case, in which case only a read lock will be needed. This will add a constant cost, but will
	// allow many more parallel threads.
	t.mutex.RLock()
	if t.testPresent(o, to) {
		treeChanged = false
		o.setPosition(to)
	}
	t.mutex.RUnlock()
	if treeChanged {
		t.mutex.Lock()
		t.remove(o)
		o.setPosition(to)
		t.checkExpand(to)
		t.add(o)
		t.mutex.Unlock()
	}
}

// Adds all objects in this or its descendants to the specified set
func (t *quadtree) collectObjects(os *[]Object) {
	if t.hasChildren {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				t.children[x][y].collectObjects(os)
			}
		}
	} else {
		*os = append(*os, t.objects...)
	}
}

// Find all objects within radius "dist" from "pos".
func (t *quadtree) findNearObjects(pos Twof, dist float64, objList *[]Object) {
	dist2 := dist * dist
	if !t.hasChildren {
		for _, o := range t.objects {
			if computeDist2(pos, o.getCurrentPosition()) > dist2 {
				continue // This object was too far away
			}
			*objList = append(*objList, o)
		}
	} else {
		// Traverse all sub squares that are inside the distance.
		for x := 0; x < 2; x++ {
			if x == 0 {
				if pos[0]-dist > t.center[0] {
					continue
				}
			} else if pos[0]+dist < t.center[0] {
				continue
			}
			for y := 0; y < 2; y++ {
				if y == 0 {
					if pos[1]-dist > t.center[1] {
						continue
					}
				} else if pos[1]+dist < t.center[1] {
					continue
				}
				t.children[x][y].findNearObjects(pos, dist, objList)
			}
		}
	}
}

// FindNearObjects finds all objects within radius 'dist' from positin 'pos'.
func (t *Quadtree) FindNearObjects(pos Twof, dist float64) []Object {
	var objList []Object
	t.mutex.RLock()
	t.findNearObjects(pos, dist, &objList)
	t.mutex.RUnlock()
	return objList
}

// Do full tree search for an object, not based on position. Used for debugging purposes.
func (this *quadtree) searchForObject(obj Object) *quadtree {
	if this == nil {
		return nil
	}
	if !this.hasChildren {
		for _, o := range this.objects {
			if o == obj {
				return this
			}
		}
	} else {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				ret := this.children[x][y].searchForObject(obj)
				if ret != nil {
					return ret
				}
			}
		}
	}
	return nil
}
