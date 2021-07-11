package main

import (
	"math/rand"
	"time"
)

const (
	MaxLevel = 16
	P        = 0.5
)

type skipListNode struct {
	item Item
	next []*skipListNode
}

type skipListOC struct {
	head  *skipListNode
	level int
}

func newSkipListOC() *skipListOC {
	rand.Seed(time.Now().UTC().UnixNano())
	return &skipListOC{
		head: &skipListNode{
			next: []*skipListNode{nil},
		},
		level: 1,
	}
}

func (o *skipListOC) Get(key string) (string, bool) {
	x := o.head

	for i := o.level; i >= 1; i-- {
		for x.next[i-1] != nil && x.next[i-1].item.Key < key {
			x = x.next[i-1]
		}
	}
	x = x.next[0]
	if x != nil && x.item.Key == key {
		return x.item.Value, true
	}
	return "", false
}

func (o *skipListOC) Put(key, value string) bool {
	// when the search is complete (and we are ready to perform the splice),
	// update[i] contains a pointer to the rightmost node of level i or
	// higher that is to the left of the location of the insertion/deletion.
	update := make([]*skipListNode, MaxLevel)
	x := o.firstGE(key, update)

	// update
	if x != nil && x.item.Key == key {
		x.item.Value = value
	} else {
		// create
		lvl := randomLevel()
		if lvl > o.level {
			for i := o.level + 1; i <= lvl; i++ {
				update[i-1] = o.head
			}
			o.level = lvl
		}

		newNode := &skipListNode{
			next: make([]*skipListNode, lvl),
			item: Item{
				Key:   key,
				Value: value,
			},
		}

		for i := 1; i <= lvl; i++ {
			if len(update[i-1].next) >= i {
				newNode.next[i-1] = update[i-1].next[i-1]
			}

			if len(update[i-1].next) < i {
				update[i-1].next = append(update[i-1].next, newNode)
			} else {
				update[i-1].next[i-1] = newNode
			}
		}
	}

	return true
}

func (o *skipListOC) Delete(key string) bool {
	// when the search is complete (and we are ready to perform the splice),
	// update[i] contains a pointer to the rightmost node of level i or
	// higher that is to the left of the location of the insertion/deletion.
	update := make([]*skipListNode, MaxLevel)
	x := o.firstGE(key, update)

	if x == nil || x.item.Key != key {
		return false
	}

	// found the item that needs to be deleted
	for i := 1; i <= o.level; i++ {
		if len(update[i-1].next) >= i {
			if update[i-1].next[i-1] == x {
				// does x necessarily have i - 1 levels?
				update[i-1].next[i-1] = x.next[i-1]
			}
		}
	}

	// update list max level
	for o.head.next[o.level-1] == nil {
		o.level -= 1
	}

	return true
}

func (o *skipListOC) firstGE(key string, update []*skipListNode) *skipListNode {
	x := o.head
	for i := o.level; i >= 1; i-- {
		for x.next[i-1] != nil && x.next[i-1].item.Key < key {
			x = x.next[i-1]
		}
		if update != nil {
			update[i-1] = x
		}
	}

	x = x.next[0]

	return x
}

func (o *skipListOC) RangeScan(startKey, endKey string) Iterator {
	node := o.firstGE(startKey, nil)
	return &skipListOCIterator{o, node, startKey, endKey}
}

type skipListOCIterator struct {
	o                *skipListOC
	node             *skipListNode
	startKey, endKey string
}

func (iter *skipListOCIterator) Next() {
	iter.node = iter.node.next[0]
}

func (iter *skipListOCIterator) Valid() bool {
	return iter.node != nil && iter.node.item.Key <= iter.endKey
}

func (iter *skipListOCIterator) Key() string {
	return iter.node.item.Key
}

func (iter *skipListOCIterator) Value() string {
	return iter.node.item.Value
}

func randomLevel() int {
	v := 1
	for rand.Float64() < P && v < MaxLevel {
		v = v + 1
	}
	return v
}
