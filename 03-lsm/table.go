package table

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"../skip_list"
)

type Item struct {
	Key, Value string
}

const (
	MAX_BLOCK_SIZE    = 4096
	KEY_LENGTH_SIZE   = 4
	VALUE_LENGTH_SIZE = 4
)

type indexEntry struct {
	key       string
	offset    uint32
	blockSize uint32
	itemCount uint32
}

/*
file format:
data_block data_block ... data_block
index_entry index_entry ... index_entry
index_offset index_entry_#

data_block format:
key_size, key, value_size, value

index_entry format:
key_size, key, offset, block_size
*/

// Given a sorted list of key/value pairs, write them out according to the format you designed.
func Build(path string, sortedItems []Item) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	buf := new(bytes.Buffer)

	totalBytesWritten := 0
	footer := []indexEntry{}
	itemCount := 0
	var lastWrittenKey string

	for _, item := range sortedItems {
		// this block if full. need to flush, clean up, and start a new one
		if buf.Len() > MAX_BLOCK_SIZE {
			// flush to file
			bytesWritten, writeErr := flushBlockToFile(f, buf)
			if writeErr != nil {
				return writeErr
			}
			// set up index entry
			footer = append(footer, indexEntry{
				key:       lastWrittenKey,
				offset:    uint32(totalBytesWritten),
				blockSize: uint32(bytesWritten),
				itemCount: uint32(itemCount),
			})
			totalBytesWritten += bytesWritten
			itemCount = 0
		}

		if item.Key == "dbofierrjmj" {
			fmt.Printf("I exist!\n")
		}

		// put bytes for this item in the byteArr for future write
		keyBytes := []byte(item.Key)
		valBytes := []byte(item.Value)

		keySizeBytes := make([]byte, KEY_LENGTH_SIZE)
		binary.BigEndian.PutUint32(keySizeBytes, uint32(len(keyBytes)))

		buf.Write(keySizeBytes)
		buf.Write(keyBytes)

		valSizeBytes := make([]byte, VALUE_LENGTH_SIZE)
		binary.BigEndian.PutUint32(valSizeBytes, uint32(len(valBytes)))

		buf.Write(valSizeBytes)
		buf.Write(valBytes)
		itemCount++
		lastWrittenKey = item.Key
		fmt.Printf("item count %d\n", itemCount)
	}

	if buf.Len() > 0 {
		bytesWritten, writeErr := flushBlockToFile(f, buf)
		if writeErr != nil {
			return writeErr
		}
		// set up index entry
		footer = append(footer, indexEntry{
			key:       lastWrittenKey,
			offset:    uint32(totalBytesWritten),
			blockSize: uint32(bytesWritten),
			itemCount: uint32(itemCount),
		})
		totalBytesWritten += bytesWritten
	}

	buf.Reset()

	// write footer to the file
	for _, entry := range footer {
		keySizeBytes := make([]byte, KEY_LENGTH_SIZE)
		binary.BigEndian.PutUint32(keySizeBytes, uint32(len(entry.key)))

		buf.Write(keySizeBytes)
		buf.WriteString(entry.key)

		offsetBytes := make([]byte, KEY_LENGTH_SIZE)
		binary.BigEndian.PutUint32(offsetBytes, entry.offset)
		buf.Write(offsetBytes)

		sizeBytes := make([]byte, KEY_LENGTH_SIZE)
		binary.BigEndian.PutUint32(sizeBytes, entry.blockSize)
		buf.Write(sizeBytes)

		countBytes := make([]byte, KEY_LENGTH_SIZE)
		binary.BigEndian.PutUint32(countBytes, entry.itemCount)
		buf.Write(countBytes)
	}

	// flush footer bytes to file
	footerBytesWritten, writeErr := f.Write(buf.Bytes())
	if writeErr != nil {
		return writeErr
	}
	log.Printf("Written %d footer bytes\n", footerBytesWritten)

	// write index_offset
	if err = binary.Write(f, binary.BigEndian, uint32(totalBytesWritten)); err != nil {
		return err
	}
	// write index_entry_#
	if err = binary.Write(f, binary.BigEndian, uint32(len(footer))); err != nil {
		return err
	}

	return nil
}

// A Table provides efficient access into sorted key/value data that's organized according
// to the format you designed.
//
// Although a Table shouldn't keep all the key/value data in memory, it should contain
// some metadata to help with efficient access (e.g. size, index, optional Bloom filter).
type Table struct {
	BlockIndex *skip_list.SkipListOC
	FilePath   string
}

// Prepares a Table for efficient access. This will likely involve reading some metadata
// in order to populate the fields of the Table struct.
func LoadTable(path string) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileReader := bufio.NewReader(f)

	f.Seek(-8, io.SeekEnd)
	buf := make([]byte, 4)

	if _, err = io.ReadFull(fileReader, buf); err != nil {
		return nil, err
	}

	indexOffset := binary.BigEndian.Uint32(buf)
	log.Printf("Index offset: %d", indexOffset)

	if _, err = io.ReadFull(fileReader, buf); err != nil {
		return nil, err
	}
	numberOfIndexEntries := int(binary.BigEndian.Uint32(buf))
	log.Printf("Index entry #: %d", numberOfIndexEntries)

	table := Table{
		BlockIndex: skip_list.NewSkipListOC(),
		FilePath:   path,
	}

	f.Seek(int64(indexOffset), io.SeekStart)

	for i := 0; i < numberOfIndexEntries; i++ {
		entry, readIndexErr := readIndexEntry(f)
		if readIndexErr != nil {
			return nil, readIndexErr
		}
		fmt.Printf("Index entry %v %d\n", entry.key, entry.offset)
		table.BlockIndex.Put(entry.key, fmt.Sprintf("%v-%v-%v", strconv.Itoa(int(entry.offset)), strconv.Itoa(int(entry.blockSize)), strconv.Itoa(int(entry.itemCount))))
	}

	return &table, nil
}

func (t *Table) Get(key string) (string, bool, error) {

	fmt.Printf("Looking for %v\n", key)

	// find the index block where the key might be
	indexNode := t.BlockIndex.FirstGE(key, nil)
	if indexNode == nil {
		return "", false, nil
	}

	valueParts := strings.Split(indexNode.Item.Value, "-")
	offset, _ := strconv.Atoi(valueParts[0])
	size, _ := strconv.Atoi(valueParts[1])
	count, _ := strconv.Atoi(valueParts[2])

	log.Printf("Offset %d Size %d Count %d \n", offset, size, count)

	f, err := os.Open(t.FilePath)
	if err != nil {
		return "", false, nil
	}
	defer f.Close()

	blockBuf := make([]byte, size)
	if _, readAtErr := f.ReadAt(blockBuf, int64(offset)); readAtErr != nil {
		return "", false, readAtErr
	}

	items := deserializeBlock(blockBuf, count)
	for i := 0; i < count; i++ {
		if items[i].Key == key {
			return items[i].Value, true, nil
		}
	}
	return "", false, nil
}

func (t *Table) RangeScan(startKey, endKey string) (Iterator, error) {
	return nil, nil
}

type Iterator interface {
	// Advances to the next item in the range. Assumes Valid() == true.
	Next()

	// Indicates whether the iterator is currently pointing to a valid item.
	Valid() bool

	// Returns the Item the iterator is currently pointing to. Assumes Valid() == true.
	Item() Item
}

func flushBlockToFile(f *os.File, buffer *bytes.Buffer) (int, error) {
	bytesWritten, writeErr := f.Write(buffer.Bytes())
	if writeErr != nil {
		return 0, writeErr
	}

	log.Printf("Written %d bytes\n", bytesWritten)

	// start a new block
	buffer.Reset()
	return bytesWritten, nil
}

func deserializeBlock(blockBuf []byte, count int) []Item {
	index := uint32(0)
	items := make([]Item, count)
	for i := 0; i < count; i++ {
		keySize := binary.BigEndian.Uint32(blockBuf[index : index+4])
		index += 4
		keyEnd := index + keySize
		key := string(blockBuf[index:keyEnd])
		index += keySize
		valSize := binary.BigEndian.Uint32(blockBuf[index : index+4])
		index += 4
		valEnd := index + valSize
		val := string(blockBuf[index:valEnd])
		index += valSize
		items[i] = Item{
			Key:   key,
			Value: val,
		}
	}
	for _, item := range items {
		fmt.Printf("Key: %v Value %v\n", item.Key, item.Value)
	}
	return items
}

func readIndexEntry(reader io.Reader) (*indexEntry, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	keySize := binary.BigEndian.Uint32(buf)

	buf = make([]byte, keySize)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	key := string(buf)

	buf = make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	indexOffset := binary.BigEndian.Uint32(buf)

	buf = make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	blockSize := binary.BigEndian.Uint32(buf)

	buf = make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	itemCount := binary.BigEndian.Uint32(buf)

	return &indexEntry{
		key:       key,
		offset:    indexOffset,
		blockSize: blockSize,
		itemCount: itemCount,
	}, nil
}
