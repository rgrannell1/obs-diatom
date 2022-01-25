package diatom

import (
	"fmt"
	"sync"
)

type Stats struct {
	Data map[string]int
	Lock sync.Locker
}

func NewStats() *Stats {
	return &Stats{
		Data: map[string]int{},
		Lock: &sync.Mutex{},
	}
}

func (stat *Stats) Add(key string) {
	stat.Lock.Lock()

	stat.Data[key]++
	fmt.Println(stat.Data)

	stat.Lock.Unlock()
}
