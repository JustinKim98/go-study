package main

import "fmt"

// #61: Propagating an inappropriate context -----------------------------------------------------------------------------------------

//Code Example: Mistake

func handler(w http.ResponseWriter, r *http.Request) {
    response, err := doSomeTask(r.Context(), r) // performs task to create a HTTP response
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    go func() { // publish the response to kafka by creating a new goroutine
        err := publish(r.Context(), response)
        // Do something with err (context may be canceled prematurely)
    }()
    writeResponse(w, response) // writes the response to the HTTP response writer
}

// If the response is written after the Kafka publication, okay fine
// If the repsonse is written before or during the Kafka publication, the message shouldn't be published


// How to avoid the mistake: call publish with an empty context
func handler(w http.ResponseWriter, r *http.Request) {
    response, err := doSomeTask(r.Context(), r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
  
    go func() {
        err := publish(context.Background(), response) // empty context
        // Do something with err
    }()
    writeResponse(w, response)
}


// #62: Starting a goroutine without knowing when to stop it -------------------------------------------------------------------------

// Goroutines are resources and can leak if not properly stopped, leading to memory exhaustion or dangling operations.
// Always provide a mechanism (e.g., context cancellation or done channel) to signal termination.
// Use defer to ensure cleanup and wait for completion using sync.WaitGroup if needed.

// Code Example: Mistake and how to avoid it

func main() {
    w := newWatcher()
    // defer w.close() // if we don't close the watcher, the goroutine will run forever
    // Run the application
}

type watcher struct {
    // Internal fields...
}

func newWatcher() watcher {
    w := watcher{}
    go w.watch() // creates a goroutine that watches external configuration
    return w
}

func (w watcher) watch() {
    // Watching logic, listen for done signal...
}

func (w watcher) close() {
    // Signal stop and close resources
}



// #63: Not being careful with goroutines and loop variables -------------------------------------------------------------------------

// Code Example: Mistake

func mistake() {
	s := []int{1, 2, 3}
	for _, i := range s {
		go func() {
			fmt.Println(i)
		}()
	}
}

// expectation: 1, 2, 3
// actual: 3, 3, 3 or 2, 3, 3 or 1, 3, 3 idk man

func kindOfCorrect() {
	s := []int{1, 2, 3}
	for _, i := range s {
		val := i
		go func() {
			fmt.Println(val)
		}()
	}
}
// expectation: 1, 2, 3
// actual: 1, 2, 3

func correct() {
	s := []int{1, 2, 3}
	for _, i := range s {
		go func(val int) { // executes a function that takes an integer as an argument
			fmt.Println(val)
		}(i) // calls this function and passes the current value of i
	}
}

// #64: Expecting Deterministic Behavior Using Select and Channels ---------------------------------------------------------------------

// When multiple cases in a select are ready, one is chosen randomly to prevent starvation, leading to non-deterministic outcomes.
// Assuming order can cause bugs like missing messages.
// Use unbuffered channels for synchronization or redesign to avoid reliance on order (e.g., close channels instead of signal channels).

// Code Example: Mistake

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
        fmt.Println("disconnection, return")
        return // May return before all messages if select chooses randomly
    }
}

// How to avoid the mistake:

messageCh := make(chan int)

go func() {
    for i := 0; i < 10; i++ {
        messageCh <- i
    }
    close(messageCh) // Use close for deterministic drain
}()

for v := range messageCh {
    fmt.Println(v)
}

// #65: Not Using Notification Channels -----------------------------------------------------------------------------------------------

// For pure signals (no data), use chan struct{} instead of chan bool to avoid ambiguity about boolean values.
// chan struct{} is idiomatic for notifications and uses zero memory allocation for signals.

// Code Example: Mistake
done := make(chan bool)
go func() {
    // Work...
    done <- true // What does true mean?
}()
<-done

// How to avoid the mistake:
done := make(chan struct{})
go func() {
    // Work...
    close(done) // Clear signal
}()
<-done


// #66: Not Using Nil Channels ------------------------------------------------------------------------------------------------------

// Code Example: How to avoid the mistake

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

ch := make(chan int) // Unbuffered
ch <- 1 // Blocks forever if no receiver

// How to avoid the mistake:

ch := make(chan int, 1) // Size 1 for non-blocking signal
ch <- 1 // Doesn't block
v := <-ch


// #68: Forgetting About Possible Side Effects with String Formatting ---------------------------------------------------------------------

// fmt functions like Sprintf call String() or Error() methods, which may have side effects or be unsafe concurrently (e.g., modifying state).
// Can cause data races or deadlocks if methods access shared resources.
// Solution: Ensure methods like String() are thread-safe or avoid formatting in concurrent contexts.

// Code Example: Mistake (Data race)

type myType struct {
    state int // Shared
}

func (m *myType) String() string {
    m.state++ // Side effect, race if formatted concurrently
    return fmt.Sprintf("%d", m.state)
}

func main() {
    mt := &myType{}
    go fmt.Println(mt) // Calls String() concurrently
    go fmt.Println(mt)
}

// How to avoid:

func (m *myType) String() string {
    return fmt.Sprintf("%d", m.state) // No side effect
}

// #69: Creating data races with append ----------------------------------------------------------------------------------------------

// append is not thread-safe; concurrent appends to the same slice cause data races on the slice header.
// Even if append returns a new slice, concurrent modifications lead to corruption.
// Solution: Use mutex protection or channels for safe concurrent updates.


// Code Example: Mistake

var s []int
var mu sync.Mutex

go func() {
    s = append(s, 1) // Race on s
}()
go func() {
    s = append(s, 2)
}()


// How to avoid:

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


// #70: Using Mutexes Inaccurately with Slices and Maps --------------------------------------------------------------------------------

// Mutexes must protect entire operations on slices/maps; partial locking leads to races.
// For maps, concurrent read/write is undefined; use sync.Map for concurrent access.
// For slices, locking only for append allows concurrent reads to see inconsistent state.
// Solution: Lock for reads and writes, or use thread-safe alternatives.

// Code Example: Mistake

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


// How to avoid:

var sm sync.Map

go func() {
    sm.Store("key", 1)
}()
go func() {
    v, _ := sm.Load("key")
    fmt.Println(v)
}()