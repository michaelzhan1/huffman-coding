package compress

import (
	"fmt"
	"strings"

	"github.com/michaelzhan1/priority-queue-go/priorityqueue"
)

// node is a node in a huffman tree
type node struct {
	value byte
	left  *node
	right *node
}

// String returns a string representation of a node
func (v *node) String() string {
	var visit func(*node, []bool) string
	visit = func(v *node, pipeDepth []bool) string {
		var tab strings.Builder
		for _, usePipe := range pipeDepth {
			if usePipe {
				tab.WriteString("\u2502 ")
			} else {
				tab.WriteString("  ")
			}
		}

		branchPrefix := tab.String() + "\u251c\u2500"
		bendPrefix := tab.String() + "\u2514\u2500"

		if v == nil {
			return bendPrefix + "nil\n"
		}

		var value string
		if v.value > 0 { // TODO: better handle byte 0
			value = string(v.value)
		}

		pipeDepth = append(pipeDepth, true)
		left := visit(v.left, pipeDepth)
		pipeDepth[len(pipeDepth)-1] = false
		right := visit(v.right, pipeDepth)
		pipeDepth = pipeDepth[:len(pipeDepth)-1]

		var sb strings.Builder
		fmt.Fprintf(&sb, branchPrefix+"Value: %q\n", value)
		fmt.Fprintf(&sb, branchPrefix+"Left:\n%s", left)
		fmt.Fprintf(&sb, bendPrefix+"Right:\n%s", right)

		return sb.String()
	}

	return visit(v, []bool{})
}

// huffTree is a huffman tree
type huffTree struct {
	root *node
}

// treeFromFreq returns a *huffTree from a frequency map
func treeFromFreq(freq map[byte]int) (*huffTree, error) {
	// TODO: add tie breaking on the priority queue ordering
	type item struct {
		weight int
		node   *node
	}

	pq := priorityqueue.NewPriorityQueue(func(a, b *item) bool { return a.weight < b.weight })
	for b, count := range freq {
		pq.Push(&item{
			weight: count,
			node:   &node{value: b},
		})
	}

	var root *node
	for !pq.IsEmpty() {
		n1, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildCompressTree)
		}

		if pq.IsEmpty() {
			root = n1.node
			break
		}

		n2, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildCompressTree)
		}

		parent := &item{
			weight: n1.weight + n2.weight,
			node: &node{
				left:  n2.node,
				right: n1.node,
			},
		}

		pq.Push(parent)
	}

	return &huffTree{root: root}, nil
}

// treeFromCode returns a *huffTree from a huffman code
// TODO: add validation for malformed codes
func treeFromCode(codes map[byte]*Code) *huffTree {
	// build tree
	// TODO: check correctness
	root := &node{}
	for b, code := range codes {
		cur := root
		for i := range code.Length {
			if (code.Bits[i/8]>>(7-i%8))&1 == 0 {
				if cur.left == nil {
					cur.left = &node{}
				}
				cur = cur.left
			} else {
				if cur.right == nil {
					cur.right = &node{}
				}
				cur = cur.right
			}
		}

		cur.value = b
	}

	return &huffTree{root: root}
}

// String returns a string representation of the tree
func (tree *huffTree) String() string {
	return tree.root.String()
}

// getCode returns a huffman code mapping
func (tree *huffTree) getCodeMap() map[byte]*Code {
	res := map[byte]*Code{}

	var visit func(*node, []bool)
	visit = func(v *node, path []bool) {
		if v == nil {
			return
		}

		// leaf
		if v.left == nil && v.right == nil {
			code := &Code{
				Bits:   [MaxCodeBytes]byte{},
				Length: len(path), // TODO: will fail for a single char file
			}

			idx := 0 // TODO: just use i?
			for i, bit := range path {
				if bit {
					code.Bits[idx] |= 1 << (7 - i%8)
				}
				if (i+1)%8 == 0 {
					idx++
				}
			}

			res[v.value] = code
			return
		}

		// traverse: false/0=left, true/1=right
		path = append(path, false)
		visit(v.left, path)
		path[len(path)-1] = true
		visit(v.right, path)
		path = path[:len(path)-1]
	}

	visit(tree.root, []bool{})
	return res
}

// newDecoder returns a closure that tracks state through the tree
func (tree *huffTree) newDecoder() func(bit bool) (byte, bool, error) {
	cur := tree.root

	return func(bit bool) (byte, bool, error) {
		if bit {
			cur = cur.right
		} else {
			cur = cur.left
		}

		if cur == nil {
			return 0, false, errCodeNotFound
		}

		if cur.left == nil && cur.right == nil {
			value := cur.value
			cur = tree.root
			return value, true, nil
		}

		return 0, false, nil
	}
}
