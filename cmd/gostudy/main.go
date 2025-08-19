package main

import (
	"fmt"
	// This controls the maxprocs environment variable in container runtimes.
	// see https://martin.baillie.id/wrote/gotchas-in-the-go-network-packages-defaults/#bonus-gomaxprocs-containers-and-the-cfs
)

func main() {
	fmt.Println("Running simple performance benchmark...")
	// SimpleBenchmark()
	CountBenchmark()
}
