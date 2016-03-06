package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

// var testFileSizes = []int{0, 100, 200, 500, 1024, 7666, 1 << 20, 1 << 24}
var testFileSizes = []int{1 << 20}

func TestFile(t *testing.T) {
	rnd := rand.New(rand.NewSource(23))

	for i, filesize := range testFileSizes {
		f := NewFile(23, filesize, 23)
		buf, err := f.ReadAll()
		if err != nil {
			t.Errorf("test %d: error %v", i, err)
			continue
		}

		if len(buf) != filesize {
			t.Errorf("test %d: invalid number of bytes returned, want %d, got %d", i, filesize, len(buf))
			continue
		}

		buf2 := make([]byte, filesize)
		n, err := f.ReadAt(buf2, 0)

		if err != nil {
			t.Errorf("test %d: error %v", i, err)
			continue
		}

		if n != filesize {
			t.Errorf("test %d: invalid number of bytes returned, want %d, got %d", i, filesize, n)
			continue
		}

		if !bytes.Equal(buf, buf2) {
			t.Errorf("test %d: wrong bytes returned", i)
			fmt.Printf("test %d: wrong bytes returned\n", i)
			continue
		}

		if filesize == 0 {
			n, err = f.ReadAt(buf2, 23)
			if n != 0 {
				t.Errorf("test %d: to read 0 bytes from empty file, got %d", i, n)
			}

			if err == nil {
				t.Errorf("test %d: expected error for empty file at offset 23 not found", i)
			}

			continue
		}

		for j := 0; j < 10; j++ {
			fmt.Printf("\n=========================================\n")

			o := rnd.Intn(filesize)
			l := rnd.Intn(filesize - o)

			fmt.Printf("  filesize %v, len %v, off %v\n", filesize, l, o)

			readatBuf := make([]byte, l)
			n, err = f.ReadAt(readatBuf, int64(o))
			fmt.Printf("n %v, l %v, o %v\n", n, l, o)
			if err != nil {
				t.Errorf("test %d/%d: reading len %v bytes at offset %v failed: %v", i, j, l, o, err)
				continue
			}

			if n != l {
				t.Errorf("test %d/%d: want %d bytes, got %d", i, j, l, n)
				fmt.Printf("test %d/%d: want %d bytes, got %d\n", i, j, l, n)
				continue
			}

			if !bytes.Equal(readatBuf, buf[o:o+l]) {
				fmt.Printf("------------------ test %d/%d: wrong bytes returned at offset %v, len %v\n", i, j, o, l)
				// fmt.Printf("  want: %02x\n", buf[o:o+l])
				// fmt.Printf("   got: %02x\n", readatBuf)
				t.Errorf("test %d/%d: wrong bytes returned at offset %v, len %v", i, j, o, l)
			}
		}
	}
}

func TestRandReader(t *testing.T) {
	rd := newRandReader(rand.New(rand.NewSource(23)))

	buf1 := make([]byte, 200)
	n, err := rd.Read(buf1)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(buf1) {
		t.Fatalf("not enough bytes returned, want %d, got %d", len(buf1), n)
	}

	for _, l := range []int{1, 2, 3, 4, 5, 6, 7, 20, 22} {
		buf2 := make([]byte, l)
		rd := newRandReader(rand.New(rand.NewSource(23)))

		for i := 0; i < 2; i++ {
			n, err := rd.Read(buf2)
			if err != nil {
				t.Error(err)
				continue
			}

			if n != len(buf2) {
				t.Errorf("not enough bytes returned, want %d, got %d", len(buf1), n)
				continue
			}

			if !bytes.Equal(buf1[i*l:(i+1)*l], buf2) {
				t.Errorf("bytes not equal: want %02x, got %02x", buf1[i*l:(i+1)*l], buf2)
			}
		}
	}
}
