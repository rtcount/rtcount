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
		Localcache_setint(k, v, 120)
		return false
	}

	if cache_int > v {
		return true
	}

	Localcache_setint(k, v, 120)
	return false
	//set max int to local cache
}

func Localcache_cache_is_small(k string, v int64) bool {
	//return false
	cache_int := localcache_getint(k)
	if cache_int == -1 {
		//don't find the k in localcache, we should search in ssdb
		Localcache_setint(k, v, 120)
		return false
	}

	if cache_int < v {
		return true
	}

	Localcache_setint(k, v, 120)
	return false
	//set max int to local cache
}

func Localcache_set(k string, v string, expireSeconds int) {
	//return
	G_cache.Set([]byte(k), []byte(v), 3600)
}

func Localcache_del(k string) bool {
	//return
	return G_cache.Del([]byte(k))
}
func Localcache_get(k string) (value []byte, err error) {
	//return
	return G_cache.Get([]byte(k))
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

func Localcache_setint(k string, v int64, expireSeconds int) {
	//return
	v_buf := bytes.NewBuffer([]byte{})
	binary.Write(v_buf, binary.BigEndian, v)

	G_cache.Set([]byte(k), v_buf.Bytes(), expireSeconds)
}
