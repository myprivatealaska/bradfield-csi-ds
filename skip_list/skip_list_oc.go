package skip_list

import (
	"math/rand"
	"time"

	"../common"
)

const (
	MaxLevel = 16
	P        = 0.5
)

type SkipListNode struct {
	Item common.Item
	Next []*SkipListNode
}

type SkipListOC struct {
	head  *SkipListNode
	level int
}

func NewSkipListOC() *SkipListOC {
	rand.Seed(time.Now().UTC().UnixNano())
	return &SkipListOC{
		head: &SkipListNode{
			Next: []*SkipListNode{nil},
		},
		level: 1,
	}
}

func (o *SkipListOC) Get(key string) (string, bool) {
	x := o.FirstGE(key, nil)
	if x != nil && x.Item.Key == key {
		return x.Item.Value, true
	}
	return "", false
}

func (o *SkipListOC) Put(key, value string) bool {
	// when the search is complete (and we are ready to perform the splice),
	// update[i] contains a pointer to the rightmost node of level i or
	// higher that is to the left of the location of the insertion/deletion.
	update := make([]*SkipListNode, MaxLevel)
	x := o.FirstGE(key, update)

	// update
	if x != nil && x.Item.Key == key {
		x.Item.Value = value
	} else {
		// create
		lvl := randomLevel()
		if lvl > o.level {
			for i := o.level + 1; i <= lvl; i++ {
				update[i-1] = o.head
			}
			o.level = lvl
		}

		newNode := &SkipListNode{
			Next: make([]*SkipListNode, lvl),
			Item: common.Item{
				Key:   key,
				Value: value,
			},
		}

		for i := 1; i <= lvl; i++ {
			if len(update[i-1].Next) >= i {
				newNode.Next[i-1] = update[i-1].Next[i-1]
			}

			if len(update[i-1].Next) < i {
				update[i-1].Next = append(update[i-1].Next, newNode)
			} else {
				update[i-1].Next[i-1] = newNode
			}
		}
	}

	return true
}

func (o *SkipListOC) Delete(key string) bool {
	// when the search is complete (and we are ready to perform the splice),
	// update[i] contains a pointer to the rightmost node of level i or
	// higher that is to the left of the location of the insertion/deletion.
	update := make([]*SkipListNode, MaxLevel)
	x := o.FirstGE(key, update)

	if x == nil || x.Item.Key != key {
		return false
	}

	// found the Item that needs to be deleted
	for i := 1; i <= o.level; i++ {
		if len(update[i-1].Next) >= i {
			if update[i-1].Next[i-1] == x {
				// does x necessarily have i - 1 levels?
				update[i-1].Next[i-1] = x.Next[i-1]
			}
		}
	}

	// update list max level
	for o.head.Next[o.level-1] == nil {
		o.level -= 1
	}

	return true
}

func (o *SkipListOC) FirstGE(key string, update []*SkipListNode) *SkipListNode {
	x := o.head
	for i := o.level; i >= 1; i-- {
		for x.Next[i-1] != nil && x.Next[i-1].Item.Key < key {
			x = x.Next[i-1]
		}
		if update != nil {
			update[i-1] = x
		}
	}

	x = x.Next[0]

	return x
}

func (o *SkipListOC) RangeScan(startKey, endKey string) common.Iterator {
	node := o.FirstGE(startKey, nil)
	return &skipListOCIterator{o, node, startKey, endKey}
}

type skipListOCIterator struct {
	o                *SkipListOC
	node             *SkipListNode
	startKey, endKey string
}

func (iter *skipListOCIterator) Next() {
	iter.node = iter.node.Next[0]
}

func (iter *skipListOCIterator) Valid() bool {
	return iter.node != nil && iter.node.Item.Key <= iter.endKey
}

func (iter *skipListOCIterator) Key() string {
	return iter.node.Item.Key
}

func (iter *skipListOCIterator) Value() string {
	return iter.node.Item.Value
}

func randomLevel() int {
	v := 1
	for rand.Float64() < P && v < MaxLevel {
		v = v + 1
	}
	return v
}
