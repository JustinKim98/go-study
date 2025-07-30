package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// #61: Propagating an inappropriate context -----------------------------------------------------------------------------------------

//Code Example: Mistake

func handler(w http.ResponseWriter, r *http.Request) {
    response, err := doSomeTask(r.Context(), r) // performs task to create a HTTP response
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    go func() { // publish the response to kafka by creating a new goroutine
        _ = publish(r.Context(), response)
        // Do something with err (context may be canceled prematurely)
    }()
    writeResponse(w, response) // writes the response to the HTTP response writer
}

// If the response is written after the Kafka publication, okay fine
// If the repsonse is written before or during the Kafka publication, the message shouldn't be published


// How to avoid the mistake: call publish with an empty context
func handlerCorrect(w http.ResponseWriter, r *http.Request) {
    response, err := doSomeTask(r.Context(), r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
  
    go func() {
        _ = publish(context.Background(), response) // empty context
        // Do something with err
    }()
    writeResponse(w, response)
}

// Helper functions for context example
func doSomeTask(ctx context.Context, r *http.Request) (string, error) {
    return "response", nil
}

func publish(ctx context.Context, response string) error {
    return nil
}

func writeResponse(w http.ResponseWriter, response string) {
    fmt.Fprintf(w, response)
}


// #62: Starting a goroutine without knowing when to stop it -------------------------------------------------------------------------

// Goroutines are resources and can leak if not properly stopped, leading to memory exhaustion or dangling operations.
// Always provide a mechanism (e.g., context cancellation or done channel) to signal termination.
// Use defer to ensure cleanup and wait for completion using sync.WaitGroup if needed.

// Code Example: Mistake and how to avoid it

func mistake62() {
    _ = newWatcher()
    // defer w.close() // if we don't close the watcher, the goroutine will run forever
    // Run the application
    fmt.Println("Watcher created without proper cleanup - goroutine will leak")
}

func avoid62() {
    w := newWatcher()
    defer w.close() // Proper cleanup
    fmt.Println("Watcher created with proper cleanup")
}

type watcher struct {
    done chan struct{}
}

func newWatcher() *watcher {
    w := &watcher{done: make(chan struct{})}
    go w.watch() // creates a goroutine that watches external configuration
    return w
}

func (w *watcher) watch() {
    // Watching logic, listen for done signal...
    select {
    case <-w.done:
        fmt.Println("Watcher stopped")
        return
    }
}

func (w *watcher) close() {
    close(w.done) // Signal stop and close resources
}



// #63: Not being careful with goroutines and loop variables -------------------------------------------------------------------------

// Code Example: Mistake

func mistake63() {
	s := []int{1, 2, 3}
	for _, i := range s {
		go func() {
			fmt.Printf("%d\n", i)
		}()
	}
}

func avoid63() {
	s := []int{1, 2, 3}
	for _, i := range s {
		val := i
		go func() {
			fmt.Printf("%d\n", val)
		}()
	}
}
	
func avoid63_2() {
    s := []int{1, 2, 3}
	for _, i := range s {
		go func(val int) { // executes a function that takes an integer as an argument
			fmt.Printf("%d\n", val)
		}(i) // calls this function and passes the current value of i
	}	
}

// #64: Expecting Deterministic Behavior Using Select and Channels ---------------------------------------------------------------------

// When multiple cases in a select are ready, one is chosen randomly to prevent starvation, leading to non-deterministic outcomes.
// Assuming order can cause bugs like missing messages.
// Use unbuffered channels for synchronization or redesign to avoid reliance on order (e.g., close channels instead of signal channels).

// Code Example: Mistake

func mistake64() {
	fmt.Println("=== Mistake 64: Non-deterministic select behavior ===")
	messageCh := make(chan int)
	disconnectCh := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			messageCh <- i
		}
		disconnectCh <- struct{}{}
	}()

	count := 0
	for {
		select {
		case v := <-messageCh:
			fmt.Printf("Received: %d\n", v)
			count++
		case <-disconnectCh:
			fmt.Printf("Disconnection, processed %d messages\n", count)
			return // May return before all messages if select chooses randomly
		}
	}
}

func avoid64() {
	fmt.Println("=== Avoid 64: Deterministic channel draining with disconnect signal ===")
	messageCh := make(chan int)
	disconnectCh := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			messageCh <- i
		}
		disconnectCh <- struct{}{}
	}()

	for {
		select {
		case v := <-messageCh:
			fmt.Println(v)
		case <-disconnectCh:
			for { // Inner for/select
				select { // Reads the remaining messages
				case v := <-messageCh:
					fmt.Println(v)
				default:
					fmt.Println("disconnection, return")
					return
				}
			}
		}
	}
}

// #65: Not Using Notification Channels -----------------------------------------------------------------------------------------------

// For pure signals (no data), use chan struct{} instead of chan bool to avoid ambiguity about boolean values.
// chan struct{} is idiomatic for notifications and uses zero memory allocation for signals.

// Code Example: Mistake
func mistake65() {
	fmt.Println("=== Mistake 65: Using bool channel for notifications ===")
	done := make(chan bool)
	go func() {
		// Work...
		fmt.Println("Work completed")
		done <- true // What does true mean?
	}()
	<-done
	fmt.Println("Received bool signal")
}

// How to avoid the mistake:
func avoid65() {
	fmt.Println("=== Avoid 65: Using struct{} channel for notifications ===")
	done := make(chan struct{}) // only need this basically
	go func() {
		// Work...
		fmt.Println("Work completed")
		close(done) // Clear signal
	}()
	<-done
	fmt.Println("Received struct{} signal")
}


// #66: Not Using Nil Channels ------------------------------------------------------------------------------------------------------

// Code Example: How to avoid the mistake

func avoid66() {
	fmt.Println("=== Avoid 66: Using nil channels for disabling ===")
	ch1 := make(chan int)
	ch2 := make(chan int)
	
	// Send some values
	go func() {
		ch1 <- 1
		ch1 <- 2
		close(ch1)
	}()
	
	go func() {
		ch2 <- 3
		ch2 <- 4
		close(ch2)
	}()
	
	merged := merge(ch1, ch2)
	
	for v := range merged {
		fmt.Printf("Merged value: %d\n", v)
	}
}

func merge(ch1, ch2 <-chan int) <-chan int {
    ch := make(chan int, 1)

    go func() {
        for ch1 != nil || ch2 != nil {
            select {
            case v, open := <-ch1:
                if !open {
                    ch1 = nil // Disable closed channel
                    break
                }
                ch <- v
            case v, open := <-ch2:
                if !open {
                    ch2 = nil
                    break
                }
                ch <- v
            }
        }
        close(ch)
    }()

    return ch
}

// #67: Being Puzzled About Channel Size ----------------------------------------------------------------------------------------------

// Unbuffered channels (size 0) provide synchronization but block senders until received.
// Buffered channels (size >0) allow asynchronous sends up to the buffer, but large buffers can hide backpressure and lead to memory issues.
// Use size 1 for non-blocking signals; choose based on use case (sync vs async) and benchmark for performance.
// Mistake: Using arbitrary sizes without understanding semantics, leading to deadlocks or leaks.

// Code Example: Mistake

func mistake67() {
	fmt.Println("=== Mistake 67: Unbuffered channel causing deadlock ===")
	_ = make(chan int) // Unbuffered
	// This would block forever if no receiver
	// ch <- 1 // Uncomment to see deadlock
	fmt.Println("Unbuffered channel created - would block if no receiver")
}

// How to avoid the mistake:

func avoid67() {
	fmt.Println("=== Avoid 67: Buffered channel for non-blocking sends ===")
	ch := make(chan int, 1) // Size 1 for non-blocking signal
	ch <- 1 // Doesn't block
	v := <-ch
	fmt.Printf("Received value: %d\n", v)
}


// #68: Forgetting About Possible Side Effects with String Formatting ---------------------------------------------------------------------

// fmt functions like Sprintf call String() or Error() methods, which may have side effects or be unsafe concurrently (e.g., modifying state).
// Can cause data races or deadlocks if methods access shared resources.
// Solution: Ensure methods like String() are thread-safe or avoid formatting in concurrent contexts.

// Code Example: Mistake (Deadlock)

type Customer struct {
	mutex sync.RWMutex
	id    string
	age   int
}

func (c *Customer) UpdateAge(age int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if age < 0 {
		return fmt.Errorf("age should be positive for customer %v", c)
	}
	c.age = age
	return nil
}

func (c *Customer) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return fmt.Sprintf("id %s, age %d", c.id, c.age)
}

func mistake68() {
	fmt.Println("=== Mistake 68: Deadlock with String method ===")
	customer := &Customer{id: "123", age: 25}
	
	// This will cause a deadlock because UpdateAge calls String() in the error message
	// while holding a write lock, but String() tries to acquire a read lock
	err := customer.UpdateAge(-1)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// How to avoid:

type CustomerSafe struct {
	mutex sync.RWMutex
	id    string
	age   int
}

func (c *CustomerSafe) UpdateAge(age int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if age < 0 {
		// Avoid calling String() while holding the lock
		return fmt.Errorf("age should be positive for customer id %s", c.id)
	}
	c.age = age
	return nil
}

func (c *CustomerSafe) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return fmt.Sprintf("id %s, age %d", c.id, c.age)
}

func avoid68() {
	fmt.Println("=== Avoid 68: No deadlock with safe String method ===")
	customer := &CustomerSafe{id: "123", age: 25}
	
	// This won't cause a deadlock because we don't call String() while holding the lock
	err := customer.UpdateAge(-1)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// #69: Creating data races with append ----------------------------------------------------------------------------------------------

// append is not thread-safe; concurrent appends to the same slice cause data races on the slice header.
// Even if append returns a new slice, concurrent modifications lead to corruption.
// Solution: Use mutex protection or channels for safe concurrent updates.

// Code Example: Mistake

func mistake69() {
	fmt.Println("=== Mistake 69: Data race with append ===")
	var s []int
	// var mu sync.Mutex // Commented out to show the race

	go func() {
		s = append(s, 1) // Race on s
	}()
	go func() {
		s = append(s, 2)
	}()
	fmt.Printf("Slice length: %d (may be corrupted)\n", len(s))
}

// How to avoid:

func avoid69() {
	fmt.Println("=== Avoid 69: Safe append with mutex ===")
	var s []int
	var mu sync.Mutex

	go func() {
		mu.Lock()
		s = append(s, 1)
		mu.Unlock()
	}()
	go func() {
		mu.Lock()
		s = append(s, 2)
		mu.Unlock()
	}()
	fmt.Printf("Slice length: %d (safe)\n", len(s))
}


// #70: Using Mutexes Inaccurately with Slices and Maps --------------------------------------------------------------------------------

// Mutexes must protect entire operations on slices/maps; partial locking leads to races.
// For maps, concurrent read/write is undefined; use sync.Map for concurrent access.
// For slices, locking only for append allows concurrent reads to see inconsistent state.
// Solution: Lock for reads and writes, or use thread-safe alternatives.

// Code Example: Mistake

func mistake70() {
	fmt.Println("=== Mistake 70: Race condition with map access ===")
	m := make(map[string]int)
	var mu sync.Mutex

	go func() {
		mu.Lock()
		m["key"] = 1
		mu.Unlock()
	}()
	go func() {
		_ = m["key"] // Race if read without lock
	}()
	fmt.Println("Map access completed (may have race condition)")
}

// How to avoid:

func avoid70() {
	fmt.Println("=== Avoid 70: Using sync.Map for concurrent access ===")
	var sm sync.Map

	go func() {
		sm.Store("key", 1)
	}()
	go func() {
		v, _ := sm.Load("key")
		fmt.Printf("Loaded value: %v\n", v)
	}()
	fmt.Println("sync.Map access completed safely")
}