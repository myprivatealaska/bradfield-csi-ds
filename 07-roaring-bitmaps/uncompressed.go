package bitmap

const wordSize = 64

type uncompressedBitmap struct {
	data []uint64
}

func newUncompressedBitmap() *uncompressedBitmap {
	return &uncompressedBitmap{
		data: make([]uint64, 100000),
	}
}

func (b *uncompressedBitmap) Get(x uint32) bool {
	//fmt.Printf("Block: %64b\n",  b.data[x/wordSize])
	return b.data[x/wordSize]>>(wordSize-1-(x%wordSize))&1 == 1
}

func (b *uncompressedBitmap) Set(x uint32) {
	//fmt.Printf("index %v x %d", x/wordSize, x)
	b.data[x/wordSize] |= 1 << (wordSize - 1 - (x % wordSize))
	//fmt.Printf("Block: %64b\n",  b.data[x/wordSize])
}

func (b *uncompressedBitmap) Union(other *uncompressedBitmap) *uncompressedBitmap {
	data := make([]uint64, 100000)

	for i := range data {
		data[i] = b.data[i] | other.data[i]
	}
	return &uncompressedBitmap{
		data: data,
	}
}

func (b *uncompressedBitmap) Intersect(other *uncompressedBitmap) *uncompressedBitmap {
	data := make([]uint64, 100000)

	for i := range data {
		data[i] = b.data[i] & other.data[i]
	}
	return &uncompressedBitmap{
		data: data,
	}
}
