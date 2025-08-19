package main

import (
	"fmt"
	"sync"
	"time"
)

type Bar struct {
	a []int64
	b []int64
}

func sumBar(bar Bar) int64 {
	var sum int64
	for i := 0; i < len(bar.a); i++ {
		sum += bar.a[i]
	}
	return sum
}

type Foo struct {
	a int64
	b int64
}

func sumFoo(foo []Foo) int64 {
	var sum int64
	for i := 0; i < len(foo); i++ {
		sum += foo[i].a
	}
	return sum
}

var pool = sync.Pool{
	New: func() interface{} {
		return make([]Foo, 1024)
	},
}

func fooFunction() {
	foo := pool.Get().([]Foo)
	foo[0].a = 1
	foo[0].b = 2
	defer pool.Put(foo)
}

// SimpleBenchmark runs a simple performance comparison between sumFoo and sumBar
func SimpleBenchmark() {
	const size = 200000
	const iterations = 10000

	// Setup data for sumFoo
	fooSlice := make([]Foo, size)
	for i := 0; i < size; i++ {
		fooSlice[i] = Foo{a: int64(i), b: int64(i * 2)}
	}

	// Setup data for sumBar
	bar := Bar{
		a: make([]int64, size),
		b: make([]int64, size),
	}
	for i := 0; i < size; i++ {
		bar.a[i] = int64(i)
		bar.b[i] = int64(i * 2)
	}

	fmt.Printf("Performance Comparison: sumFoo vs sumBar\n")
	fmt.Printf("Dataset size: %d elements\n", size)
	fmt.Printf("Iterations: %d\n\n", iterations)

	// Benchmark sumFoo
	start := time.Now()
	var resultFoo int64
	for i := 0; i < iterations; i++ {
		resultFoo = sumFoo(fooSlice)
	}
	durationFoo := time.Since(start)

	// Benchmark sumBar
	start = time.Now()
	var resultBar int64
	for i := 0; i < iterations; i++ {
		resultBar = sumBar(bar)
	}
	durationBar := time.Since(start)

	// Display results
	fmt.Printf("sumFoo Results:\n")
	fmt.Printf("  Result: %d\n", resultFoo)
	fmt.Printf("  Total time: %v\n", durationFoo)
	fmt.Printf("  Average per operation: %v\n", durationFoo/time.Duration(iterations))

	fmt.Printf("\nsumBar Results:\n")
	fmt.Printf("  Result: %d\n", resultBar)
	fmt.Printf("  Total time: %v\n", durationBar)
	fmt.Printf("  Average per operation: %v\n", durationBar/time.Duration(iterations))

	// Performance comparison
	if durationFoo < durationBar {
		ratio := float64(durationBar) / float64(durationFoo)
		fmt.Printf("\nsumFoo is %.2fx faster than sumBar\n", ratio)
	} else if durationBar < durationFoo {
		ratio := float64(durationFoo) / float64(durationBar)
		fmt.Printf("\nsumBar is %.2fx faster than sumFoo\n", ratio)
	} else {
		fmt.Printf("\n Both functions have similar performance\n")
	}

	// Verify results are the same
	if resultFoo == resultBar {
		fmt.Printf("Both functions produce the same result: %d\n", resultFoo)
	} else {
		fmt.Printf("Results differ: sumFoo=%d, sumBar=%d\n", resultFoo, resultBar)
	}
}

type Input struct {
	a int64
	b int64
}

type Result struct {
	sumA int64
	sumB int64
}

type FastResult struct {
	sumA int64
	_    [56]byte // padding region
	sumB int64
}

func count(inputs []Input) Result {
	wg := sync.WaitGroup{}
	wg.Add(2)

	result := Result{}

	go func() {
		for i := 0; i < len(inputs); i++ {
			result.sumA += inputs[i].a
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < len(inputs); i++ {
			result.sumB += inputs[i].b
		}
		wg.Done()
	}()

	wg.Wait()
	return result
}

// countFast does the same work as count but writes into a padded FastResult
// to minimize false sharing between goroutines updating sumA and sumB.
func countFast(inputs []Input) FastResult {
	wg := sync.WaitGroup{}
	wg.Add(2)

	result := FastResult{}

	go func() {
		for i := 0; i < len(inputs); i++ {
			result.sumA += inputs[i].a
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < len(inputs); i++ {
			result.sumB += inputs[i].b
		}
		wg.Done()
	}()

	wg.Wait()
	return result
}

// CountBenchmark compares performance of count (Result) vs countFast (FastResult).
func CountBenchmark() {
	const size = 200000
	const iterations = 50000

	// Prepare input data
	inputs := make([]Input, size)
	for i := 0; i < size; i++ {
		inputs[i] = Input{a: int64(i), b: int64(i * 2)}
	}

	fmt.Printf("Count Benchmark: Result vs FastResult\n")
	fmt.Printf("Dataset size: %d elements\n", size)
	fmt.Printf("Iterations: %d\n\n", iterations)

	// Benchmark count (Result)
	start := time.Now()
	var r Result
	for i := 0; i < iterations; i++ {
		r = count(inputs)
	}
	durationResult := time.Since(start)

	// Benchmark countFast (FastResult)
	start = time.Now()
	var fr FastResult
	for i := 0; i < iterations; i++ {
		fr = countFast(inputs)
	}
	durationFast := time.Since(start)

	fmt.Printf("Result (unpadded)\n")
	fmt.Printf("  sumA: %d, sumB: %d\n", r.sumA, r.sumB)
	fmt.Printf("  Total time: %v\n", durationResult)
	fmt.Printf("  Average per operation: %v\n\n", durationResult/time.Duration(iterations))

	fmt.Printf("FastResult (padded)\n")
	fmt.Printf("  sumA: %d, sumB: %d\n", fr.sumA, fr.sumB)
	fmt.Printf("  Total time: %v\n", durationFast)
	fmt.Printf("  Average per operation: %v\n\n", durationFast/time.Duration(iterations))

	if durationFast < durationResult {
		ratio := float64(durationResult) / float64(durationFast)
		fmt.Printf("FastResult is %.2fx faster than Result\n", ratio)
	} else if durationResult < durationFast {
		ratio := float64(durationFast) / float64(durationResult)
		fmt.Printf("Result is %.2fx faster than FastResult\n", ratio)
	} else {
		fmt.Printf("Both versions have similar performance\n")
	}

	// Verify results are identical
	if r.sumA == fr.sumA && r.sumB == fr.sumB {
		fmt.Printf("Both versions produce the same sums.\n")
	} else {
		fmt.Printf("Mismatch: Result(a=%d,b=%d) vs FastResult(a=%d,b=%d)\n", r.sumA, r.sumB, fr.sumA, fr.sumB)
	}
}
