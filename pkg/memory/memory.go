// Package memory Is a slight copy of the memory storage, but far from the storage interface it can not only work with bytes
// but directly store any kind of data without having to encode it each time, which gives a huge speed advantage
package memory

import (
	"sync"
	"time"
)

type Storage struct {
	sync.RWMutex
	data  map[string]item // data
	array []interface{}
	book  map[string]int
}

type item struct {
	// max value is 4294967295 -> Sun Feb 07 2106 06:28:15 GMT+0000
	e uint32      // exp
	v interface{} // val
}

func New() *Storage {
	store := &Storage{
		data:    make(map[string]item),
		array:   []interface{}{},
		book:    make(map[string]int),
		RWMutex: sync.RWMutex{},
	}
	//utils.StartTimeStampUpdater()
	//go store.gc(1 * time.Second)
	return store
}

// Get All
func (s *Storage) GetAll() []interface{} {
	s.RLock()
	newArray := s.array
	s.RUnlock()

	return newArray
}

// Get All
func (s *Storage) Length() int {
	s.RLock()
	length := len(s.array)
	s.RUnlock()

	return length
}

// Get value by key
func (s *Storage) Get(key string) interface{} {
	s.RLock()
	v, ok := s.data[key]
	s.RUnlock()
	if !ok {
		return nil
	}
	// if !ok || v.e != 0 && v.e <= atomic.LoadUint32(&utils.Timestamp) {
	// 	return nil
	// }
	return v.v
}

// Set key with value
func (s *Storage) Set(key string, val interface{}, ttl time.Duration) {
	var exp uint32 = 0
	// // if ttl > 0 {
	// // 	exp = uint32(ttl.Seconds()) + atomic.LoadUint32(&utils.Timestamp)
	// // }
	i := item{exp, val}
	//	s.Lock()
	s.data[key] = i
	s.array = append(s.array, val)
	s.book[key] = len(s.array) - 1
	// s.Unlock()
}

// Delete key by key
func (s *Storage) Delete(key string) {
	s.Lock()
	i := s.book[key]
	s.array = append(s.array[:i], s.array[i+1:]...)
	delete(s.data, key)
	delete(s.book, key)
	for k := range s.book {
		if s.book[k] > i {
			s.book[k] = s.book[k] - 1
		}
	}
	s.Unlock()
}

// Swap key with another key
func (s *Storage) Swap(oldKey, key string) {
	s.Lock()
	s.data[key] = s.data[oldKey]
	delete(s.data, oldKey)
	s.Unlock()
}

// Reset all keys
func (s *Storage) Reset() {
	nd := make(map[string]item)
	s.Lock()
	s.data = nd
	s.Unlock()
}

// func (s *Storage) gc(sleep time.Duration) {
// 	ticker := time.NewTicker(sleep)
// 	defer ticker.Stop()
// 	var expired []string

// 	for range ticker.C {
// 		ts := atomic.LoadUint32(&utils.Timestamp)
// 		expired = expired[:0]
// 		s.RLock()
// 		for key, v := range s.data {
// 			if v.e != 0 && v.e <= ts {
// 				expired = append(expired, key)
// 			}
// 		}
// 		s.RUnlock()
// 		s.Lock()
// 		// Double-checked locking.
// 		// We might have replaced the item in the meantime.
// 		for i := range expired {
// 			v := s.data[expired[i]]
// 			if v.e != 0 && v.e <= ts {
// 				delete(s.data, expired[i])
// 			}
// 		}
// 		s.Unlock()
// 	}
// }
