package bbloom

import (
	"bufio"
	// "bytes"
	// "encoding/json"
	"fmt"
	// "hash"
	// "hash/crc64"
	// "hash/fnv"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"os"
	"testing"
	"time"
)

func Test_New(t *testing.T) {

	Convey("Given n = 512, location/hashs = 1", t, func() {
		n := 512
		locs := 1

		Convey("When created bf", func() {
			bf := New(float64(n), float64(locs))

			Convey("bf should be {{[0 ... 0] 'ptr'} 9 511 1 55}", nil)
			// So(fmt.Sprint(bf), ShouldEqual, "[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]")
			So(fmt.Sprintf("%v %v %v %v", bf.sizeExp, bf.size, bf.setLocs, bf.shift), ShouldEqual, "9 511 1 23")

			Convey("When Add([]byte('Manfred'))", func() {
				bf.Add([]byte("Manfred"))

				Convey("Has([]byte('Manfred') is true", nil)
				So(bf.Has([]byte("Manfred")), ShouldBeTrue)

				Convey("Has([]byte('manfred') is false", nil)
				So(bf.Has([]byte("manfred")), ShouldBeFalse)
			})

		})

	})

}

func Test_JSONMarschal_JSONUnmarshal(t *testing.T) {

	Convey("When created bf(64,1) and populated with 10 words and bf.JSONMarshal", t, func() {
		bf := New(float64(64), float64(1))
		w := []byte{100, 200, 130}
		for i := 0; i < 10; i++ {
			w = append(w, byte((i+1)*10))
			bf.Add(w)
		}
		Json := bf.JSONMarshal()

		Convey("bf.JSONMarschal() should be equal to JSONUnmarshal(bf.JSONMarschal())", nil)
		nbf := JSONUnmarshal(Json)
		cnt := 0
		w = []byte{100, 200, 130}
		for i := 0; i < 10; i++ {
			w = append(w, byte((i+1)*10))
			if !nbf.Has(w) {
				cnt++
			}
		}
		So(cnt, ShouldEqual, 0)
	})

}

func Test_Timings(t *testing.T) {

	Convey("When created bf(10* 2**16, 7) populated with 2**16 words", t, func() {
		n := 1 << 16
		bf := New(float64(n*10), float64(7))

		file, err := os.Open("words.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		wordlist1 := make([][]byte, n)
		for i, _ := range wordlist1 {
			if scanner.Scan() {
				wordlist1[i] = []byte(scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		Convey("When created bf(n*20,3): Time to populate with n words", func() {
			st := time.Now()
			repeats := int64(100)
			for r := int64(0); r < repeats; r++ {
				bf := New(float64(n*10), float64(7))
				for i, _ := range wordlist1 {
					bf.Add(wordlist1[i])
				}
			}
			tm := time.Since(st).Nanoseconds()
			fmt.Printf("Create bloom filter (n*20,3) and 'Add' n=%v items to bloom filter (%v repetitions): %v ns (%v ns/op)\n", n, repeats, tm/repeats, tm/(int64(n)*repeats))
		})
		Convey("Time to check n words", func() {
			st := time.Now()
			repeats := int64(100)
			for r := int64(0); r < repeats; r++ {
				for i, _ := range wordlist1 {
					bf.Has(wordlist1[i])
				}
			}
			tm := time.Since(st).Nanoseconds()
			fmt.Printf("Bloom filter 'Has' %v items (%v repetitions): %v ns (%v ns/op)\n", n, repeats, tm/repeats, tm/(int64(n)*repeats))
		})
		Convey("Time to JSONMarshal bloom filter", func() {
			st := time.Now()
			repeats := int64(20)
			for r := int64(0); r < repeats; r++ {
				_ = bf.JSONMarshal()
			}
			tm := time.Since(st).Seconds()
			fmt.Printf("JSONMarshal %v times: %v s (%v s/op)", repeats, tm, tm/float64(repeats))
		})
	})

}

func Test_bf_Distributions(t *testing.T) {
	Convey("Test for wrong positivs in bf if there're 3 fields for each loc: len(bf)=9*2<<16, 3 Locs/entry. wordlist length = 2<<16 ", t, func() {
		n := 1 << 16
		bf := New(float64(n), float64(0.01))

		file, err := os.Open("words.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		wordlist1 := make([][]byte, n)
		for i, _ := range wordlist1 {
			if scanner.Scan() {
				wordlist1[i] = []byte(scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		Convey("When populated bf with wordlist1", func() {

			for _, w := range wordlist1 {
				bf.Add(w)
			}

			Convey("bf.Has(word) should be positiv for all words in wordlist1", nil)
			notIn := 0
			for _, w := range wordlist1 {
				if !bf.Has(w) {
					notIn++
				}
			}

			So(notIn, ShouldEqual, 0)

			Convey("Wrong positivs < 0.1%: bf.Has(word) should be negativ for more than 99.9% of words not in wordlist1", nil)

			file, err := os.Open("words.txt")
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			cnt := 0
			wrongIn := 0
			scanner := bufio.NewScanner(file)
			for i := 0; i < n; i++ {
				if scanner.Scan() {
					_ = []byte(scanner.Text())
				}
			}
			for scanner.Scan() {
				if bf.Has([]byte(scanner.Text())) {
					wrongIn++
				}
				cnt++
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}

			So(0.04, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			So(0.02, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			So(0.01, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			So(0.005, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			So(0.001, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			// So(0.0005, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			// So(0.0001, ShouldBeGreaterThan, float64(wrongIn)/float64(cnt))
			log.Println(wrongIn, cnt)
		})

	})

	wordlist := make([][]byte, 1<<16)
	for i := range wordlist {
		wordlist[i] = []byte(fmt.Sprintf("2014/08/01 %v %v", time.Now().Nanosecond(), i))
	}
	log.Println(string(wordlist[0]))

	bl := New(64.0, 1.0)
	st := time.Now()
	for r := 0; r < 100; r++ {
		bl = New(65536.0, 0.01)
		for i := range wordlist {
			bl.Add(wordlist[i])
		}
	}
	tm := time.Since(st).Nanoseconds()
	log.Println("Add", tm, "ns", tm/int64(100*65536), "ns/op")

	st = time.Now()
	for r := 0; r < 100; r++ {
		for i := range wordlist {
			bl.Has(wordlist[i])
		}
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("Has", tm, "ns", tm/int64(100*65536), "ns/op")

	cnt := 0
	st = time.Now()
	for r := 0; r < 100; r++ {
		for i := range wordlist {
			if !bl.Has(wordlist[i]) {
				cnt++
			}
		}
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("Has failed: ", cnt, " -- ", tm, "ns", tm/int64(100*65536), "ns/op")

	cnt = 0
	var Json []byte
	st = time.Now()
	for r := 0; r < 10; r++ {
		Json = bl.JSONMarshal()
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("JSONMarshal: ", tm, "ns", tm/int64(10), "ns/op")

	cnt = 0
	var newbl Bloom
	st = time.Now()
	for r := 0; r < 10; r++ {
		newbl = JSONUnmarshal(Json)
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("JSONUnmarshal: ", tm, "ns", tm/int64(10), "ns/op")

	cnt = 0
	st = time.Now()
	for r := 0; r < 100; r++ {
		for i := range wordlist {
			if !newbl.Has(wordlist[i]) {
				cnt++
			}
		}
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("newbl.Has failed: ", cnt, " -- ", tm, "ns", tm/int64(100*65536), "ns/op")

	for i := range wordlist {
		wordlist[i] = []byte(fmt.Sprintf("2014/08/01 %v 0%v", time.Now().Nanosecond(), i))
	}
	log.Println(string(wordlist[0]))

	cnt = 0
	st = time.Now()
	for r := 0; r < 100; r++ {
		for i := range wordlist {
			if bl.Has(wordlist[i]) {
				cnt++
			}
		}
	}
	tm = time.Since(st).Nanoseconds()
	log.Println("Wrong positives: ", cnt, " -- ", tm, "ns", tm/int64(100*65536), "ns/op")

	// log.Println("--------------- test with bloom (bool bloom filter) below ---------------------------------------------\n")
	// import "github.com/AndreasBriese/bloom"
	// for i := range wordlist {
	// 	wordlist[i] = []byte(fmt.Sprintf("2014/08/01 %v %v", time.Now().Nanosecond(), i))
	// }
	// log.Println(string(wordlist[0]))

	// bf := bloom.New(64.0, 1.0)
	// st = time.Now()
	// for r := 0; r < 100; r++ {
	// 	bf = bloom.New(65536.0, 0.01)
	// 	for i := range wordlist {
	// 		bf.Add(wordlist[i])
	// 	}
	// }
	// tm = time.Since(st).Nanoseconds()
	// log.Println("Add", tm, "ns", tm/int64(100*65536), "ns/op")

	// st = time.Now()
	// for r := 0; r < 100; r++ {
	// 	for i := range wordlist {
	// 		bf.Has(wordlist[i])
	// 	}
	// }
	// tm = time.Since(st).Nanoseconds()
	// log.Println("Has", tm, "ns", tm/int64(100*65536), "ns/op")

	// cnt = 0
	// st = time.Now()
	// for r := 0; r < 100; r++ {
	// 	for i := range wordlist {
	// 		if !bf.Has(wordlist[i]) {
	// 			cnt++
	// 		}
	// 	}
	// }
	// tm = time.Since(st).Nanoseconds()
	// log.Println("Has failed: ", cnt, " -- ", tm, "ns", tm/int64(100*65536), "ns/op")

	// for i := range wordlist {
	// 	wordlist[i] = []byte(fmt.Sprintf("2014/08/01 %v 0%v", time.Now().Nanosecond(), i))
	// }
	// log.Println(string(wordlist[0]))

	// cnt = 0
	// st = time.Now()
	// for r := 0; r < 100; r++ {
	// 	for i := range wordlist {
	// 		if bf.Has(wordlist[i]) {
	// 			cnt++
	// 		}
	// 	}
	// }
	// tm = time.Since(st).Nanoseconds()
	// log.Println("Wrong positives: ", cnt, " -- ", tm, "ns", tm/int64(100*65536), "ns/op")

}
