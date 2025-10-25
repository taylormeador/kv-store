package store

import "testing"

func BenchmarkStoreGet(b *testing.B) {
	s := NewStore()
	s.Set("key", "value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Get("key")
		}
	})
}

func BenchmarkStoreSet(b *testing.B) {
	s := NewStore()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Set("key", "value")
		}
	})
}

func BenchmarkStoreMixed(b *testing.B) {
	s := NewStore()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%5 == 0 {
				s.Set("key", "value")
			} else {
				s.Get("key")
			}
			i++
		}
	})
}
