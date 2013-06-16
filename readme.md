Quadtree
========
Maintain a list of objects.
The purpose is to keep the cost down for finding near objects.

* All objects must fulfill the 'Object' interface, with a GetPosition() and a SetPosition().
* The package is thread safe as long as the position of objects in the quadtree does not change other than when SetPosition() callback is called.

This is still work in progress.
