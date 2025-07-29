package main

import (
	"fmt"
	// "context"
	// "log/slog"
	// "os"

	// This controls the maxprocs environment variable in container runtimes.
	// see https://martin.baillie.id/wrote/gotchas-in-the-go-network-packages-defaults/#bonus-gomaxprocs-containers-and-the-cfs
	// "github.com/JustinKim98/go-study/internal/log"
	// "go.uber.org/automaxprocs/maxprocs"
)

func main() {

	// Logger configuration
	// logger := log.New(
	// 	log.WithLevel(os.Getenv("LOG_LEVEL")),
	// 	log.WithSource(),
	// )

	// Example of calling concurrency practice functions
	fmt.Println("=== Concurrency Practice Examples ===\n")
	
	// Example 63: Loop variables in goroutines
	mistake63()
	fmt.Println()
	//avoid63()
	// fmt.Println()
	
	// Example 64: Non-deterministic select behavior
	// mistake64()
	// fmt.Println()
	// avoid64()
	// fmt.Println()
	
	// Example 65: Notification channels
	// mistake65()
	// fmt.Println()
	// avoid65()
	// fmt.Println()
	
	// Example 66: Nil channels
	// avoid66()
	// fmt.Println()
	
	// Example 67: Channel size
	// mistake67()
	// fmt.Println()
	// avoid67()
	// fmt.Println()
	
	// Example 68: String formatting side effects
	// mistake68()
	// fmt.Println()
	// avoid68()
	// fmt.Println()
	
	// Example 69: Data races with append
	// mistake69()
	// fmt.Println()
	// avoid69()
	// fmt.Println()
	
	// Example 70: Mutexes with maps
	// mistake70()
	// fmt.Println()
	// avoid70()
	// fmt.Println()

	// a := [3]int{0, 1, 2}
	// for i, v := range a {
	// 	a[2] = 10
	// 	if i == 2 {
	// 		fmt.Println(v)
	// 	}
	// }

	// if err := run(logger); err != nil {
	// 	logger.ErrorContext(context.Background(), "an error occurred", slog.String("error", err.Error()))
	// 	os.Exit(1)
	// }

}

// func run(logger *slog.Logger) error {
// 	ctx := context.Background()

// 	_, err := maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
// 		logger.DebugContext(ctx, fmt.Sprintf(s, i...))
// 	}))
// 	if err != nil {
// 		return fmt.Errorf("setting max procs: %w", err)
// 	}

// 	logger.InfoContext(ctx, "Hello world!", slog.String("location", "world"))

// 	return nil
// }
