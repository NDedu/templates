package main

// Benchmarks for the inputs that used to be slow: go test -run '^$' -bench .

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkQuickSort(b *testing.B) {
	const n = 50_000
	r := rand.New(rand.NewSource(1))
	inputs := []struct {
		name  string
		value func(i int) int
	}{
		{"random", func(int) int { return r.Int() }},
		{"sorted", func(i int) int { return i }},
		{"few-distinct", func(int) int { return r.Intn(4) }},
		{"all-equal", func(int) int { return 0 }}, // was O(n^2) with the two-way partition
	}

	for _, input := range inputs {
		original := make([]int, n)
		for i := range original {
			original[i] = input.value(i)
		}
		array := make([]int, n)

		b.Run(input.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				copy(array, original)
				b.StartTimer()
				QuickSort(array)
			}
		})
	}
}

func BenchmarkHashMapPut(b *testing.B) {
	keys := make([]string, 100_000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	for i := 0; i < b.N; i++ {
		hashMap := NewHashMap(0) // starts with 16 buckets and has to grow
		for j, key := range keys {
			hashMap.Put(key, j)
		}
	}
}

func BenchmarkTreeMapPutSorted(b *testing.B) {
	keys := make([]string, 100_000)
	for i := range keys {
		keys[i] = fmt.Sprintf("%06d", i) // sorted order turned the plain BST into a linked list
	}

	for i := 0; i < b.N; i++ {
		treeMap := NewTreeMap()
		for j, key := range keys {
			treeMap.Put(key, j)
		}
	}
}
