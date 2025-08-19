package main

import (
	"testing"
)

// BenchmarkSumFoo benchmarks the sumFoo function
func BenchmarkSumFoo(b *testing.B) {
	// Create test data - slice of Foo structs
	const size = 10000
	fooSlice := make([]Foo, size)
	for i := 0; i < size; i++ {
		fooSlice[i] = Foo{
			a: int64(i),
			b: int64(i * 2),
		}
	}

	// Reset the timer to exclude setup time
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		_ = sumFoo(fooSlice)
	}
}

// BenchmarkSumBar benchmarks the sumBar function
func BenchmarkSumBar(b *testing.B) {
	// Create test data - Bar struct with slice of int64
	const size = 10000
	bar := Bar{
		a: make([]int64, size),
		b: make([]int64, size),
	}
	for i := 0; i < size; i++ {
		bar.a[i] = int64(i)
		bar.b[i] = int64(i * 2)
	}

	// Reset the timer to exclude setup time
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		_ = sumBar(bar)
	}
}

// BenchmarkSumFooSmall benchmarks sumFoo with smaller dataset
func BenchmarkSumFooSmall(b *testing.B) {
	const size = 100
	fooSlice := make([]Foo, size)
	for i := 0; i < size; i++ {
		fooSlice[i] = Foo{
			a: int64(i),
			b: int64(i * 2),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sumFoo(fooSlice)
	}
}

// BenchmarkSumBarSmall benchmarks sumBar with smaller dataset
func BenchmarkSumBarSmall(b *testing.B) {
	const size = 100
	bar := Bar{
		a: make([]int64, size),
		b: make([]int64, size),
	}
	for i := 0; i < size; i++ {
		bar.a[i] = int64(i)
		bar.b[i] = int64(i * 2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sumBar(bar)
	}
}

// BenchmarkSumFooLarge benchmarks sumFoo with larger dataset
func BenchmarkSumFooLarge(b *testing.B) {
	const size = 100000
	fooSlice := make([]Foo, size)
	for i := 0; i < size; i++ {
		fooSlice[i] = Foo{
			a: int64(i),
			b: int64(i * 2),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sumFoo(fooSlice)
	}
}

// BenchmarkSumBarLarge benchmarks sumBar with larger dataset
func BenchmarkSumBarLarge(b *testing.B) {
	const size = 100000
	bar := Bar{
		a: make([]int64, size),
		b: make([]int64, size),
	}
	for i := 0; i < size; i++ {
		bar.a[i] = int64(i)
		bar.b[i] = int64(i * 2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sumBar(bar)
	}
}

