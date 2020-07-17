package sniper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/lotsa"
)

func TestPack(t *testing.T) {
	addr := 1<<26 - 5
	size := byte(5)
	some32 := addrSizeMarshal(uint32(addr), size)
	s, l := addrSizeUnmarshal(some32)
	if s != uint32(addr) || l != 32 {
		t.Errorf("get addr = %d, size=%d", s, l)
	}
	addr = 1<<28 - 1
	size = byte(19)
	maxAddrSize := addrSizeMarshal(uint32(addr), size)
	s, l = addrSizeUnmarshal(maxAddrSize)
	if s != uint32(addr) || l != 524288 {
		t.Errorf("get addr = %d, size=%d ", s, l)
	}
}

func TestHashCol(t *testing.T) {
	//println(1 << 32)
	k2 := make([]byte, 8)
	binary.BigEndian.PutUint64(k2, uint64(16_123_243))
	k3 := make([]byte, 8)
	binary.BigEndian.PutUint64(k3, uint64(106_987_520))
	//println(hash(k2), hash(k3))
	//mgdbywinfo uzmqkfjche 720448991
	//println("str", hash([]byte("mgdbywinfo")), hash([]byte("uzmqkfjche")))
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

	s, err := Open(Dir("1"))
	assert.NoError(t, err)

	err = s.Set([]byte("hello"), []byte("go"))
	assert.NoError(t, err)

	err = s.Set([]byte("hello"), []byte("world"))
	assert.NoError(t, err)

	res, err := s.Get([]byte("hello"))
	assert.NoError(t, err)

	assert.Equal(t, true, bytes.Equal(res, []byte("world")))

	assert.Equal(t, 1, s.Count())

	err = s.Set([]byte("ahello"), []byte("aworld"))
	assert.NoError(t, err)

	err = s.Set([]byte("bhello"), []byte("bworld"))
	assert.NoError(t, err)

	err = s.Set([]byte("chello"), []byte("cworld"))
	assert.NoError(t, err)

	var kvb bool

	kls := []string{"hello", "ahello", "bhello", "chello"}
	vls := []string{"world", "aworld", "bworld", "cworld"}
	var kfs []string
	var vfs []string

	walk := func(key []byte, val []byte) bool {
		kfs = append(kfs, string(key))
		vfs = append(vfs, string(val))
		return false
	}

	err = s.Walk(walk)
	if err != nil {
		t.Errorf("Walk error: %v", err)
	}

	kvb = IsEqual(kls, kfs)
	if !kvb { t.Errorf("Walk keys arrays not equal: %v, %v", kls, kfs) }
	kvb = IsEqual(vls, vfs)
	if !kvb { t.Errorf("Walk values arrays not equal: %v, %v", vls, vfs) }

	_, err = s.Delete([]byte("ahello"))
	assert.NoError(t, err)

	_, err = s.Delete([]byte("bhello"))
	assert.NoError(t, err)

	_, err = s.Delete([]byte("chello"))
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

	s, err = Open(Dir("1"))
	assert.NoError(t, err)

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
	sniperBench(seed())
}

func randKey(rnd *rand.Rand, n int) []byte {
	s := make([]byte, n)
	rnd.Read(s)
	for i := 0; i < n; i++ {
		s[i] = 'a' + (s[i] % 26)
	}
	return s
}

func seed() ([][]byte, int) {
	seed := int64(1570109110136449000) //time.Now().UnixNano() //1570108152262917000
	// println(seed)
	rng := rand.New(rand.NewSource(seed))
	N := 100_000
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
	return keys, N
}

func sniperBench(keys [][]byte, N int) {
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true

	fmt.Println("-- sniper --")
	DeleteStore("1")
	//s, err := Open(Dir("1"),SyncInterval(1*time.Second)) Игорь, попробуй вот так
	s, err := Open(Dir("1"))
	if err != nil {
		panic(err)
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	fmt.Printf("Alloc = %v MiB Total = %v MiB\n", (ms.Alloc / 1024 / 1024), (ms.TotalAlloc / 1024 / 1024))

	fmt.Print("set: ")
	coll := 0
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(i))
		//println("set", i, keys[i], b)
		err := s.Set(keys[i], b)
		if err == ErrCollision {
			coll++
			err = nil
		}
		if err != nil {
			panic(err)
		}
	})
	runtime.ReadMemStats(&ms)

	fmt.Printf("Alloc = %v MiB Total = %v MiB Coll=%d\n", (ms.Alloc / 1024 / 1024), (ms.TotalAlloc / 1024 / 1024), coll)
	coll = 0
	fmt.Print("get: ")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		b, err := s.Get(keys[i])
		if err != nil {
			println("errget", string(keys[i]))
			panic(err)
		}
		v := binary.BigEndian.Uint64(b)

		if uint64(i) != v {
			println("get error:", string(keys[i]), i, v)
			panic("bad news")
		}
	})

	runtime.ReadMemStats(&ms)

	fmt.Printf("Alloc = %v MiB Total = %v MiB\n", (ms.Alloc / 1024 / 1024), (ms.TotalAlloc / 1024 / 1024))

	fmt.Print("del: ")
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		s.Delete(keys[i])
	})
	err = DeleteStore("1")
	if err != nil {
		panic("bad news")
	}
}

func IsEqual(a1 []string, a2 []string) bool {
	sort.Strings(a1)
	sort.Strings(a2)
	if len(a1) == len(a2) {
		for i, v := range a1 {
			if v != a2[i] {
				return false
			}
		}
	} else {
		return false
	}
	return true
}
