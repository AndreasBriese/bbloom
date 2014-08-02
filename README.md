## bbloom: a bitset Bloom filter for go/golang
===

package implements a fast bloom filter with real 'bitset' and JSONMarshal/JSONUnmarshal to store/reload the Bloom filter. 

NOTE: the package uses unsafe.Pointer to set and read the bits from the bitset. If you're uncomfortable with using the unsafe package, please consider using my bloom filter package at github.com/AndreasBriese/bloom

===

This bloom filter was developed to strengthen a website-log database and was tested and optimized for this log-entry mask: "2014/%02i/%02i %02i:%02i:%02i /info.html". 
Nonetheless bbloom should work with any other form of entries. 

Hash function is Berkeley DB smdb hash (slightly modified to optimize for smaller bitsets len<=4096). smdb <--- http://www.cse.yorku.ca/~oz/hash.html

Minimum hashset size is: 512 ([4]uint64; will be set automatically). 

###install

```sh
go get github.com/AndreasBriese/bbloom
```

###test
+ change to folder ../bloom 
+ create wordlist in file "words.txt" (you might use `go run wordlister.go `)
+ run 'go test' within the folder

```go
go test
```

If you've installed the GOCONVEY TDD-framework http://goconvey.co/ you can run the tests automatically.

### usage

after installation add

```go
import (
	...
	"github.com/AndreasBriese/bbloom"
	...
	)
```

at your header. In the program use

```go
// create a bloom filter for 65536 items and 1 % wrong-positive ratio 
bf := bbloom.New(float64(1<<16), float64(0.01))

// or 
// create a bloom filter with 650000 for 65536 items and 7 locs per hash explicitly
// bf = bbloom.New(float64(650000), float64(7))
// or
bf = bbloom.New(650000.0, 7.0)

// add one item
bf.Add([]byte("butter"))

// check if item is in the filter
isIn := bf.Has([]byte("butter"))    // should be true
isNotIn := bf.Has([]byte("Butter")) // should be false

// convert to JSON ([]byte) 
Json := bf.JSONMarshal()

// restore a bloom filter from storage 
bf = JSONUnmarshal(Json)
```

to work with the bloom filter.

### why 'fast'? 

It's about 3 times faster than William Fitzgeralds bitset bloom filter https://github.com/willf/bloom . And it is about so fast as my []bool set variant for Boom filters (see https://github.com/AndreasBriese/bloom ): 

	
	Bloom filter (filter size 524288, 7 hashlocs)
	github.com/AndreasBriese/bbloom 'Add' 65536 items (10 repetitions): 6595800 ns (100 ns/op)
    github.com/AndreasBriese/bbloom 'Has' 65536 items (10 repetitions): 5986600 ns (91 ns/op)
	github.com/AndreasBriese/bloom 'Add' 65536 items (10 repetitions): 6304684 ns (96 ns/op)
	github.com/AndreasBriese/bloom 'Has' 65536 items (10 repetitions): 6568663 ns (100 ns/op)
	
	github.com/willf/bloom 'Add' 65536 items (10 repetitions): 24367224 ns (371 ns/op)
	github.com/willf/bloom 'Test' 65536 items (10 repetitions): 21881142 ns (333 ns/op)
	github.com/dataence/bloom/standard 'Add' 65536 items (10 repetitions): 23041644 ns (351 ns/op)
	github.com/dataence/bloom/standard 'Check' 65536 items (10 repetitions): 19153133 ns (292 ns/op)
	github.com/cabello/bloom 'Add' 65536 items (10 repetitions): 131921507 ns (2012 ns/op)
	github.com/cabello/bloom 'Contains' 65536 items (10 repetitions): 131108962 ns (2000 ns/op)

(on MBPro15 OSX10.8.5 i7 4Core 2.4Ghz)


With 32bit bloom filters (bloom32) using smdb, bloom32 does hashing with only 2 bit shifts, one xor and one substraction per byte. smdb is about as fast as fnv64a but gives less collisions with the dataset (see mask above). bloom.New(float64(10 * 1<<16),float64(7)) populated with 1<<16 random items from the dataset (see above) and tested against the rest results in less than 0.05% collisions.   
