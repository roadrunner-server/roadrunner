package bst

// BST ...
type BST struct {
	// registered topic, not unique
	topic string
	// associated connections with the topic
	uuids map[string]struct{}

	// left and right subtrees
	left  *BST
	right *BST
}

func NewBST() Storage {
	return &BST{
		uuids: make(map[string]struct{}, 10),
	}
}

// Insert uuid to the topic
func (b *BST) Insert(uuid string, topic string) {
	curr := b

	for {
		if topic == curr.topic {
			curr.uuids[uuid] = struct{}{}
			return
		}
		// if topic less than curr topic
		if topic < curr.topic {
			if curr.left == nil {
				curr.left = &BST{
					topic: topic,
					uuids: map[string]struct{}{uuid: {}},
				}
				return
			}
			// move forward
			curr = curr.left
		} else {
			if curr.right == nil {
				curr.right = &BST{
					topic: topic,
					uuids: map[string]struct{}{uuid: {}},
				}
				return
			}

			curr = curr.right
		}
	}
}

func (b *BST) Contains(topic string) bool {
	curr := b
	for curr != nil {
		switch {
		case topic < curr.topic:
			curr = curr.left
		case topic > curr.topic:
			curr = curr.right
		case topic == curr.topic:
			return true
		}
	}

	return false
}

func (b *BST) Get(topic string) map[string]struct{} {
	curr := b
	for curr != nil {
		switch {
		case topic < curr.topic:
			curr = curr.left
		case topic > curr.topic:
			curr = curr.right
		case topic == curr.topic:
			return curr.uuids
		}
	}

	return nil
}

func (b *BST) Remove(uuid string, topic string) {
	b.removeHelper(uuid, topic, nil)
}

func (b *BST) removeHelper(uuid string, topic string, parent *BST) { //nolint:gocognit
	curr := b
	for curr != nil {
		if topic < curr.topic { //nolint:gocritic
			parent = curr
			curr = curr.left
		} else if topic > curr.topic {
			parent = curr
			curr = curr.right
		} else {
			// if more than 1 topic - remove only topic, do not remove the whole vertex
			if len(curr.uuids) > 1 {
				if _, ok := curr.uuids[uuid]; ok {
					delete(curr.uuids, uuid)
					return
				}
			}

			if curr.left != nil && curr.right != nil { //nolint:gocritic
				curr.topic, curr.uuids = curr.right.traverseForMinString()
				curr.right.removeHelper(curr.topic, uuid, curr)
			} else if parent == nil {
				if curr.left != nil { //nolint:gocritic
					curr.topic = curr.left.topic
					curr.uuids = curr.left.uuids

					curr.right = curr.left.right
					curr.left = curr.left.left
				} else if curr.right != nil {
					curr.topic = curr.right.topic
					curr.uuids = curr.right.uuids

					curr.left = curr.right.left
					curr.right = curr.right.right
				} else { //nolint:staticcheck
					// single node tree
				}
			} else if parent.left == curr {
				if curr.left != nil {
					parent.left = curr.left
				} else {
					parent.left = curr.right
				}
			} else if parent.right == curr {
				if curr.left != nil {
					parent.right = curr.left
				} else {
					parent.right = curr.right
				}
			}
			break
		}
	}
}

//go:inline
func (b *BST) traverseForMinString() (string, map[string]struct{}) {
	if b.left == nil {
		return b.topic, b.uuids
	}
	return b.left.traverseForMinString()
}
