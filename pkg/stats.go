package diatom

import (
	"fmt"
	"sync"
)

type Stats struct {
	Data map[string]int
	Lock sync.Locker
}

func (stat *Stats) Add(key string) {
	stat.Lock.Lock()

	stat.Data[key]++
	fmt.Println(stat.Data)

	stat.Lock.Unlock()
}
