package sniper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"runtime"
	"testing"

	"github.com/spaolacci/murmur3"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/lotsa"
)

func TestPack(t *testing.T) {
	addr := 1<<26 - 5
	size := byte(5)
	some32 := addrsizePack(uint32(addr), size)
	s, l := addrsizeUnpack(some32)
	if s != uint32(addr) || l != 32 {
		t.Errorf("get addr = %d, size=%d want 60000,4", s, l)
	}
	addr = 1<<28 - 1
	size = byte(19)
	maxAddrSize := addrsizePack(uint32(addr), size)
	s, l = addrsizeUnpack(maxAddrSize)
	if s != uint32(addr) || l != 524288 {
		t.Errorf("get addr = %d, size=%d ", s, l)
	}
}

func TestHashCol(t *testing.T) {
	//get: k:[110 111 104 101 111 111 109 122 121 122] key:[105 113 101 118 119 122 113 116 99 116] val: [0 0 0 0 0 6 157 19]
	//k1 := []byte{110, 111, 104, 101, 111, 111, 109, 122, 121, 122}
	//k2 := []byte{105, 113, 101, 118, 119, 122, 113, 116, 99, 11}
	println(1 << 32)
	k2 := make([]byte, 8)
	binary.BigEndian.PutUint64(k2, uint64(16_123_243))
	k3 := make([]byte, 8)
	binary.BigEndian.PutUint64(k3, uint64(106_987_520))
	println(hash(k2), hash(k3))
	//mgdbywinfo uzmqkfjche 720448991
	println("str", hash([]byte("mgdbywinfo")), hash([]byte("uzmqkfjche")))
	//		 4_294_967_296
	sizet := 1_000_000
	m := make(map[uint32]int, sizet)
	for i := 0; i < sizet; i++ {
		k1 := make([]byte, 8)
		binary.BigEndian.PutUint64(k1, uint64(i))
		h := hash(k1)
		if _, ok := m[h]; ok {
			println("collision", h, i, m[h])
			break
		}
		m[h] = i
	}

}
func TestPower(t *testing.T) {

	p, v := NextPowerOf2(256)
	if p != 8 || v != 256 {
		t.Errorf("get p = %d,v=%d want 8,256", p, v)
	}

	p, v = NextPowerOf2(1023)
	if p != 10 || v != 1024 {
		t.Errorf("get p = %d,v=%d want 10,1024", p, v)
	}

	p, v = NextPowerOf2(4294967294) //2^32-1-1
	if p != 32 || v != 4294967295 {
		t.Errorf("get p = %d,v=%d want 33,4294967295", p, v)
	}

	p, v = NextPowerOf2(3)
	if p != 2 || v != 4 {
		t.Errorf("get p = %d,v=%d want 2,4", p, v)
	}
	p, v = NextPowerOf2(0)
	if p != 0 || v != 0 {
		t.Errorf("get p = %d,v=%d want 0,0", p, v)
	}
}

func TestCmd(t *testing.T) {
	err := DeleteStore("1")
	assert.NoError(t, err)

	s, err := Open("1")

	assert.NoError(t, err)
	err = s.Set([]byte("hello"), []byte("go"))
	assert.NoError(t, err)

	err = s.Set([]byte("hello"), []byte("world"))
	assert.NoError(t, err)

	res, err := s.Get([]byte("hello"))
	assert.NoError(t, err)

	assert.Equal(t, true, bytes.Equal(res, []byte("world")))

	assert.Equal(t, 1, s.Count())

	err = s.Close()
	s, err = Open("1")
	res, err = s.Get([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, true, bytes.Equal(res, []byte("world")))
	assert.Equal(t, 1, s.Count())

	deleted, err := s.Delete([]byte("hello"))
	assert.NoError(t, err)
	assert.True(t, deleted)
	assert.Equal(t, 0, s.Count())

	counter := []byte("counter")

	cnt, err := s.Incr(counter, uint64(1))
	assert.NoError(t, err)
	assert.Equal(t, 1, int(cnt))
	cnt, err = s.Incr(counter, uint64(42))
	assert.NoError(t, err)
	assert.Equal(t, 43, int(cnt))

	cnt, err = s.Decr(counter, uint64(2))
	assert.NoError(t, err)
	assert.Equal(t, 41, int(cnt))

	//overflow
	cnt, err = s.Decr(counter, uint64(42))
	assert.NoError(t, err)
	assert.Equal(t, uint64(18446744073709551615), uint64(cnt))

	err = s.Backup()
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

	err = DeleteStore("1")
	assert.NoError(t, err)

}

func randKey(rnd *rand.Rand, n int) []byte {
	s := make([]byte, n)
	rnd.Read(s)
	for i := 0; i < n; i++ {
		s[i] = 'a' + (s[i] % 26)
	}
	return s
}

func TestGet(t *testing.T) {
	s, err := Open("1")
	if err != nil {
		panic(err)
	}
	v, err := s.Get([]byte("uzmqkfjche"))
	println(err.Error(), v)
	s.Close()
}
func TestSniperSpeed(t *testing.T) {

	seed := int64(1570109110136449000) //time.Now().UnixNano() //1570108152262917000
	// println(seed)
	rng := rand.New(rand.NewSource(seed))
	N := 10_000_000
	K := 10

	fmt.Printf("\n")
	fmt.Printf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("\n")
	fmt.Printf("     number of cpus: %d\n", runtime.NumCPU())
	fmt.Printf("     number of keys: %d\n", N)
	fmt.Printf("            keysize: %d\n", K)
	fmt.Printf("        random seed: %d\n", seed)

	fmt.Printf("\n")

	keysm := make(map[string]bool, N)
	for len(keysm) < N {
		keysm[string(randKey(rng, K))] = true
	}
	keys := make([][]byte, 0, N)
	for key := range keysm {
		keys = append(keys, []byte(key))
	}

	lotsa.Output = os.Stdout
	lotsa.MemUsage = true

	println("-- murmur3 --")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		murmur3.Sum32WithSeed(keys[i], 0)
	})

	println("-- fnv1 --")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		hash32a := fnv.New32()
		hash32a.Write(keys[i])
		hash32a.Sum32()
	})

	println("-- sniper --")
	DeleteStore("1")
	s, _ := Open("1")
	print("set: ")
	coll := 0
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(i))
		//println("set", i, keys[i], b)
		err := s.Set(keys[i], b)
		if err == errCollision {
			coll++
			err = nil
		}
		if err != nil {
			panic(err)
		}
	})
	println("setcol:", coll)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc = %v MiB Total = %v MiB\n", (ms.Alloc / 1024 / 1024), (ms.TotalAlloc / 1024 / 1024))
	coll = 0
	print("get: ")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		b, err := s.Get(keys[i])
		if err != nil {
			println("errget", string(keys[i]))
			panic(err)
		}
		v := binary.BigEndian.Uint64(b)

		if uint64(i) != v {
			println("bad news:", string(keys[i]), i, v)
			panic("bad news")
		}
	})
	println("getcol:", coll)

	print("del: ")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		//s.Delete(keys[i])
	})
	//DeleteStore("1")
	println()

	//uncomment for badger test
	/*
		DeleteStore("badger_test")
		bd, err := newBadgerdb("badger_test")
		if err != nil {
			panic(err)
		}
		println("-- badger --")
		print("set: ")

		lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
			txn := bd.NewTransaction(true) // Read-write txn
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, uint64(i))

			err = txn.SetEntry(badger.NewEntry(keys[i], b))
			if err != nil {
				log.Fatal(err)
			}
			err = txn.Commit()
			if err != nil {
				log.Fatal(err)
			}

		})

		print("get: ")
		lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
			var val []byte
			err := bd.View(func(txn *badger.Txn) error {
				item, err := txn.Get(keys[i])
				if err != nil {
					return err
				}
				val, err = item.ValueCopy(val)
				return err
			})
			if err != nil {
				log.Fatal(err)
			}
			v := binary.BigEndian.Uint64(val)
			if uint64(i) != v {
				panic("bad news")
			}
		})

		print("del: ")
		lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
			txn := bd.NewTransaction(true)
			err := txn.Delete(keys[i])
			if err != nil {
				log.Fatal(err)
			}
			err = txn.Commit()
			if err != nil {
				log.Fatal(err)
			}
		})

		DeleteStore("badger_test")
	*/
}

/*
func newBadgerdb(path string) (*badger.DB, error) {

	os.MkdirAll(path, os.FileMode(0777))
	opts := badger.DefaultOptions(path)
	opts.SyncWrites = false
	return badger.Open(opts)
}*/