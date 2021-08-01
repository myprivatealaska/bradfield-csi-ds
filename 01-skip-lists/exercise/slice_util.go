package main

import (
	"sort"

	"../../common"
)

// This file contains helper functions used by both the "slice" and the
// "linked block" approaches.

// Find the first index i such that items[i].Key >= key
// returns len(items) if no such index exists
func sliceFirstGE(items []common.Item, key string) int {
	// Use the binary search implementation from the standard library
	return sort.Search(len(items), func(i int) bool {
		return items[i].Key >= key
	})

	// If we wanted to use linear search instead, we could do this:
	// i := 0
	// for i < len(items) && items[i].Key < key {
	//     i++
	// }
	// return i
}

func sliceGet(items []common.Item, key string) (string, bool) {
	i := sliceFirstGE(items, key)
	if i < len(items) && items[i].Key == key {
		return items[i].Value, true
	}
	return "", false
}

func slicePut(items *[]common.Item, key, value string) bool {
	i := sliceFirstGE(*items, key)
	if i == len(*items) {
		*items = append(*items, common.Item{key, value})
		return true
	} else if (*items)[i].Key == key {
		(*items)[i].Value = value
		return false
	} else {
		var newItems []common.Item
		newItems = append(newItems, (*items)[:i]...)
		newItems = append(newItems, common.Item{key, value})
		newItems = append(newItems, (*items)[i:]...)
		*items = newItems
		return true
	}
}

func sliceDelete(items *[]common.Item, key string) bool {
	i := sliceFirstGE(*items, key)
	if i < len(*items) && (*items)[i].Key == key {
		*items = append((*items)[:i], (*items)[i+1:]...)
		return true
	}
	return false
}
