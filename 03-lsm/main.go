package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"sort"
)

// min and max are inclusive.
func randomWord(min, max int) string {
	n := min + rand.Intn(max-min+1)
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		c := rune(rand.Intn(26))
		buf.WriteRune('a' + c)
	}
	return buf.String()
}

// All items are guaranteed to have unique keys.
func generateSortedItems(n int) []Item {
	m := make(map[string]struct{}, n)
	for len(m) < n {
		key := randomWord(8, 16)
		m[key] = struct{}{}
	}
	keys := make([]string, 0, n)
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Item, n)
	for i, key := range keys {
		value := randomWord(10, 20)
		result[i] = Item{key, value}
	}
	return result
}

func main() {
	dir, err := ioutil.TempDir("", "table")
	if err != nil {
		panic(err)
	}
	// Clean up temp directory at end of test; you can remove this for debugging.
	//defer os.RemoveAll(dir)

	tmpfile := filepath.Join(dir, "tmpfile")

	n := 1000
	sortedItems := generateSortedItems(n)

	toInclude := sortedItems[:n/2]
	log.Println(toInclude[200])
	//toExclude := sortedItems[n/2:]

	err = Build(tmpfile, toInclude)
	if err != nil {
		panic(err)
	}

	table, err := LoadTable(tmpfile)
	if err != nil {
		panic(err)
	}

	val, _, err := table.Get("fgfkstuhmhyu")
	if err != nil {
		panic(err)
	}
	log.Println(val)

	//xqjdjuzgfvmnhdmb
}
