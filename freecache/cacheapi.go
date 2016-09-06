package freecache

import (
	"bytes"
	"encoding/binary"
)

var G_cache *Cache

func Localcache_cache_init(size int) {

	G_cache = NewCache(size)
}

func Localcache_check_and_set(k string) bool {
	//return false
	err := G_cache.Set([]byte(k), []byte("1"), 120)
	if err == KeyFound {
		return true
	}
	return false
}

func Localcache_cache_is_big(k string, v int64) bool {
	//return false
	cache_int := localcache_getint(k)
	if cache_int == -1 {
		//don't find the k in localcache, we should search in ssdb
		localcache_setint(k, v)
		return false
	}

	if cache_int > v {
		return true
	}

	localcache_setint(k, v)
	return false
	//set max int to local cache
}

func Localcache_cache_is_small(k string, v int64) bool {
	//return false
	cache_int := localcache_getint(k)
	if cache_int == -1 {
		//don't find the k in localcache, we should search in ssdb
		localcache_setint(k, v)
		return false
	}

	if cache_int < v {
		return true
	}

	localcache_setint(k, v)
	return false
	//set max int to local cache
}

func localcache_set(k string, v string) {
	//return
	G_cache.Set([]byte(k), []byte(v), 120)
}

func localcache_getint(k string) int64 {
	//return -1
	value, err := G_cache.Get([]byte(k))
	if err == ErrNotFound {
		return -1
	}

	var x int64
	b_buf := bytes.NewBuffer(value)

	binary.Read(b_buf, binary.BigEndian, &x)

	return x
}

func localcache_setint(k string, v int64) {
	//return
	v_buf := bytes.NewBuffer([]byte{})
	binary.Write(v_buf, binary.BigEndian, v)

	G_cache.Set([]byte(k), v_buf.Bytes(), 120)
}
