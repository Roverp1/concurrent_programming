package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SharedData struct {
	value *int
	mu    sync.Mutex
	cond  *sync.Cond

	lastWriter string
	lastReader string
}

func main() {
	fmt.Printf("=== Lab 1 ===\n")
	fmt.Printf("Starting 4 goroutines for 30 secons...\n")

	const STUDENT_ID = "121546"

	sharedData := NewSharedData()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var waitGroup sync.WaitGroup

	waitGroup.Add(4)

	go producerT1(ctx, sharedData, &waitGroup, fmt.Sprintf("%s#Writer#1", STUDENT_ID))
	go producerT2(ctx, sharedData, &waitGroup, fmt.Sprintf("%s#Writer#2", STUDENT_ID))
	go consumerT3(ctx, sharedData, &waitGroup, fmt.Sprintf("%s#Reader#1", STUDENT_ID))
	go monitorT4(ctx, sharedData, &waitGroup, fmt.Sprintf("%s#Monitor#1", STUDENT_ID))

	waitGroup.Wait()

	fmt.Printf("\n=== All goroutines finished ===\n")
}

func NewSharedData() *SharedData {
	sd := &SharedData{}

	sd.cond = sync.NewCond(&sd.mu)
	return sd
}

func (sd *SharedData) WriteValue(value int, writerName string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	for sd.value != nil {
		sd.cond.Wait()
	}

	sd.value = &value
	sd.lastWriter = writerName

	fmt.Printf("[%s] wrote: %d\n", writerName, value)

	sd.cond.Broadcast()
}

func (sd *SharedData) ReadValue(readerName string) (int, bool) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	for sd.value == nil {
		sd.cond.Wait()
	}

	readValue := *sd.value
	sd.value = nil
	sd.lastReader = readerName

	return readValue, true
}

func (sd *SharedData) GetStatus() (value *int, writer string, reader string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.value, sd.lastWriter, sd.lastReader
}

func producerT1(ctx context.Context, sd *SharedData, waitGroup *sync.WaitGroup, name string) {
	defer waitGroup.Done()

	randNumber := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] shutting down\n", name)
			return
		default:
		}

		value := randNumber.Intn(17) + 21

		sd.WriteValue(value, name)

		sleepTime := time.Duration(randNumber.Intn(1000)+500) * time.Millisecond
		time.Sleep(sleepTime)
	}
}

func producerT2(ctx context.Context, sd *SharedData, waitGroup *sync.WaitGroup, name string) {
	defer waitGroup.Done()

	randNumber := rand.New(rand.NewSource(time.Now().UnixNano() + 1))

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] shutting down...\n", name)
			return
		default:
		}

		value := randNumber.Intn(2864) + 1337

		sd.WriteValue(value, name)

		sleepTime := time.Duration(randNumber.Intn(1000)+500) * time.Millisecond
		time.Sleep(sleepTime)
	}
}

func consumerT3(ctx context.Context, sd *SharedData, waitGroup *sync.WaitGroup, name string) {
	defer waitGroup.Done()

	sum := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] shutting down... final sum: %d\n", name, sum)
			return
		default:
		}

		resultChan := make(chan int, 1)

		go func() {
			value, _ := sd.ReadValue(name)
			resultChan <- value
		}()

		select {
		case value := <-resultChan:
			sum += value
			fmt.Printf("[%s] current sum: %d\n", name, sum)
		case <-ctx.Done():
			fmt.Printf("[%s] shutting down... final sum: %d\n", name, sum)
			return
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func monitorT4(ctx context.Context, sd *SharedData, waitGroup *sync.WaitGroup, name string) {
	defer waitGroup.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] shutting down...\n", name)
			return
		case <-ticker.C:
			value, lastWriter, lastReader := sd.GetStatus()

			fmt.Printf("\n=== [%s] STATUS REPORT ===\n", name)
			if value != nil {
				fmt.Printf("Shared value: %d\n", *value)
			} else {
				fmt.Printf("Shared value: nil\n")
			}
			fmt.Printf("Last writer: %s\n", lastWriter)
			fmt.Printf("Last Reader: %s\n", lastReader)
			fmt.Printf("All goroutines: running\n")
			fmt.Printf("========================\n\n")
		}
	}
}
