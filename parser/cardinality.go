package parser

import (
	"errors"
	h "github.com/eclesh/hyperloglog"
	"hash/fnv"
)

type CardinalityCounter struct {
	m   uint
	hll map[string]*h.HyperLogLog
}

func NewCardinalityCounter(m uint) *CardinalityCounter {
	c := &CardinalityCounter{m: m}
	c.hll = make(map[string]*h.HyperLogLog)
	return c
}

func (this *CardinalityCounter) Add(key string, data interface{}) (err error) {
	if _, ok := this.hll[key]; !ok {
		this.hll[key], err = h.New(this.m)
	}

	switch data.(type) {
	case string:
		hash := fnv.New32()
		hash.Write([]byte(data.(string)))
		this.hll[key].Add(hash.Sum32())
	case int:
		this.hll[key].Add(uint32(data.(int)))
	case int16:
		this.hll[key].Add(uint32(data.(int16)))
	case int32:
		this.hll[key].Add(uint32(data.(int32)))
	case int64:
		this.hll[key].Add(uint32(data.(int64)))
	case uint:
		this.hll[key].Add(uint32(data.(uint)))
	case uint16:
		this.hll[key].Add(uint32(data.(uint16)))
	case uint32:
		this.hll[key].Add(data.(uint32))
	case uint64:
		this.hll[key].Add(uint32(data.(uint64)))
	default:
		err = errors.New("unkown type")
	}

	return
}

func (this *CardinalityCounter) Reset(key string) {
	if _, ok := this.hll[key]; !ok {
		return
	}

	this.hll[key].Reset()
}

func (this *CardinalityCounter) Count(key string) uint64 {
	if _, ok := this.hll[key]; !ok {
		return 0
	}

	return this.hll[key].Count()
}
