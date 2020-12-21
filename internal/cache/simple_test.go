package cache

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func randomData(size int) map[string]string {
	rand.Seed(0)

	m := make(map[string]string, size)
	for i := 0; i < size; i++ {
		m[strconv.Itoa(i)] = strconv.FormatUint(rand.Uint64(), 10)
	}
	return m
}

func testData(t *testing.T, size int) (*Cache, map[string]string) {
	t.Helper()
	c := New(size)
	if c == nil {
		t.Fatal("got nil cache")
	}
	assertCacheSize(t, c, 0)
	d := randomData(size)
	return c, d
}

func assertCacheSize(t *testing.T, c *Cache, want int) {
	t.Helper()
	if l := c.Len(); l != want {
		t.Errorf("expected cache with '%d' items, got '%d' items", want, l)
	}
}

func TestSimple(t *testing.T) {
	c, d := testData(t, 1000000)
	for tk, tv := range d {
		c.Set(tk, tv, 0)
		if _, found := c.m[tk]; !found {
			t.Errorf("failed setting key '%s'", tk)
		}

		v, found := c.Get(tk)
		if !found {
			t.Errorf("failed getting key '%s'", tk)
		}
		if v != tv {
			t.Errorf("got value '%s', expected '%s'", v, tv)
		}

		c.Del(tk)
		v, found = c.Get(tk)
		if found || v == tv {
			t.Errorf("expected nil key, got '%s'", v)
		}
	}
	assertCacheSize(t, c, 0)
}

func TestPassiveExpires(t *testing.T) {
	size := 1000000
	c, d := testData(t, size)
	for tk, tv := range d {
		c.Set(tk, tv, 1)
	}
	assertCacheSize(t, c, size)

	time.Sleep(1 * time.Second)
	for tk := range d {
		v, found := c.Get(tk)
		if found {
			t.Errorf("expected expired key, got '%s'", v)
		}
	}
	assertCacheSize(t, c, 0)
}

func TestActiveExpires(t *testing.T) {
	size := 1000000
	c, d := testData(t, size)
	for tk, tv := range d {
		c.Set(tk, tv, 1)
	}
	assertCacheSize(t, c, size)

	cancel := c.StartGC(250 * time.Millisecond)
	time.Sleep(1250 * time.Millisecond)
	cancel()
	assertCacheSize(t, c, 0)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkSimpleSet(b *testing.B) {
	c := New(b.N)
	d := randomData(b.N)
	b.ReportAllocs()
	b.ResetTimer()

	for k, v := range d {
		c.Set(k, v, 0)
	}
}

func BenchmarkSimpleGet(b *testing.B) {
	c := New(b.N)
	d := randomData(b.N)
	for k, v := range d {
		c.Set(k, v, 0)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for k := range d {
		c.Get(k)
	}
}

func BenchmarkSimpleDel(b *testing.B) {
	c := New(b.N)
	d := randomData(b.N)
	for k, v := range d {
		c.Set(k, v, 0)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for k := range d {
		c.Del(k)
	}
}
