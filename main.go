package main

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Philosopher struct {
	id           int
	leftFork     *sync.Mutex
	rightFork    *sync.Mutex
	lastMealTime time.Time
	isAlive      bool
	eatCount     atomic.Uint32
}

type Monitor struct {
	philosophers []*Philosopher
	mu           sync.Mutex
}

const (
	philosophersNum = 5
	maxHungerTime   = 9 * time.Second
	timeToEat       = 2 * time.Second
	thinkingTime    = 5 * time.Second
	eatNum          = 4
)

func (m *Monitor) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(500 * time.Millisecond) // Small delay to reduce CPU usage
		now := time.Now()
		m.mu.Lock()
		for _, philo := range m.philosophers {
			if philo.isAlive && now.Sub(philo.lastMealTime) > maxHungerTime {
				log.Printf("philosopher (%d) has died from hunger\n", philo.id)
				philo.isAlive = false
			}
		}
		m.mu.Unlock()
		allEated := true

		for _, philo := range m.philosophers {
			if philo.eatCount.Load() < eatNum {
				allEated = false
				break
			}
		}
		if allEated {
			log.Printf("all philosophers have eated %d times, stopping monitor\n", eatNum)
			return
		}

		// Check if all philosophers are dead
		allDead := true
		for _, philo := range m.philosophers {
			if philo.isAlive {
				allDead = false
				break
			}
		}
		if allDead {
			log.Println("all philosophers have died, stopping monitor.")
			return
		}
	}
}

func (m *Monitor) UpdateLastMeal(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id >= 0 && id < len(m.philosophers) {
		m.philosophers[id].lastMealTime = time.Now()
		m.philosophers[id].eatCount.Add(1)
	}
}

func (philo *Philosopher) Run(wg *sync.WaitGroup, monitor *Monitor) {
	defer wg.Done()
	for _ = range eatNum {
		log.Printf("philosohler (%d) is thinking\n", philo.id)
		time.Sleep(thinkingTime)
		if philo.id%2 != 0 {
			philo.leftFork.Lock()
			log.Printf("philosopher (%d) grabed left fork\n", philo.id)
			philo.rightFork.Lock()
			log.Printf("philosopher (%d) grabed right fork\n", philo.id)
		} else {
			log.Printf("philosopher (%d) grabed right fork\n", philo.id)
			philo.rightFork.Lock()
			log.Printf("philosopher (%d) grabed left fork\n", philo.id)
			philo.leftFork.Lock()
		}
		monitor.UpdateLastMeal(philo.id)
		log.Printf("philosopher (%d) is eating\n", philo.id)
		time.Sleep(timeToEat)
		philo.leftFork.Unlock()
		log.Printf("philosopher (%d) released left fork\n", philo.id)
		philo.rightFork.Unlock()
		log.Printf("philosopher (%d) released right fork\n", philo.id)
		log.Printf("philosopher (%d) finished eating\n", philo.id)
	}
}

func main() {
	var wg sync.WaitGroup

	forks := make([]sync.Mutex, philosophersNum)
	philosophers := make([]*Philosopher, philosophersNum)
	for i := range philosophersNum {
		philosophers[i] = &Philosopher{
			id:           i,
			leftFork:     &forks[i],
			rightFork:    &forks[(i+1)%philosophersNum],
			isAlive:      true,
			lastMealTime: time.Now(),
		}
	}

	monitor := Monitor{
		philosophers: philosophers,
	}

	for i := range philosophersNum {
		philosophers[i] = &Philosopher{
			id:           i,
			leftFork:     &forks[i],
			rightFork:    &forks[(i+1)%philosophersNum],
			isAlive:      true,
			lastMealTime: time.Now(),
		}
		wg.Add(1)
		go philosophers[i].Run(&wg, &monitor)
	}
	wg.Add(1)
	go monitor.Run(&wg)
	wg.Wait()
}
