package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	THINKING = 0
	HUNGRY   = 1
	EATING   = 2
	N        = 5
)

type Monitor struct {
	states [N]uint8
	conds  [N]*sync.Cond
	mu     sync.Mutex
}

type Philosopher struct {
	id      uint8
	thinks  uint16
	eats    uint16
	monitor *Monitor
	mu      sync.Mutex
}

func main() {
	monitor := NewMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phils := [N]Philosopher{}
	for i := range phils {
		phils[i].id = uint8(i)
		phils[i].thinks = 0
		phils[i].eats = 0
		phils[i].monitor = monitor
		phils[i].mu = sync.Mutex{}
	}

	var wg sync.WaitGroup
	wg.Add(N)

	fmt.Printf("Starting the goroutines\n")
	for i := range phils {
		go func(p *Philosopher) {
			defer wg.Done()
			p.run(ctx)
		}(&phils[i])
	}

	wg.Wait()

	fmt.Printf("\n==============\n")
	for i := range phils {
		fmt.Printf("Philosopher #%d - Thought: %d times, Ate %d times\n", i, phils[i].thinks, phils[i].eats)
	}
}

func NewMonitor() (monitor *Monitor) {
	monitor = new(Monitor)

	for i := range monitor.states {
		monitor.states[i] = THINKING
		monitor.conds[i] = sync.NewCond(&monitor.mu)
	}

	return
}

func (monitor *Monitor) test(i uint8) {
	left := (i + N - 1) % N
	right := (i + 1) % N

	if monitor.states[i] == HUNGRY && monitor.states[left] != EATING && monitor.states[right] != EATING {
		monitor.states[i] = EATING
		monitor.conds[i].Signal()
	}
}

func (monitor *Monitor) takeFork(i uint8) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	monitor.states[i] = HUNGRY
	monitor.test(i)
	for monitor.states[i] != EATING {
		monitor.conds[i].Wait()
	}
}

func (monitor *Monitor) putFork(i uint8) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	monitor.states[i] = THINKING
	monitor.test((i + N - 1) % N)
	monitor.test((i + 1) % N)
}

func (p *Philosopher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		p.mu.Lock()
		p.thinks++
		// fmt.Printf("Philosopher #%d is thinking\n", p.id)
		p.mu.Unlock()
		time.Sleep(50 * time.Millisecond)

		// fmt.Printf("Philosopher #%d is hungary\n", p.id)
		p.monitor.takeFork(p.id)

		p.mu.Lock()
		p.eats++
		// fmt.Printf("Philosopher #%d is eating\n", p.id)
		p.mu.Unlock()
		time.Sleep(30 * time.Millisecond)

		p.monitor.putFork(p.id)
	}
}
