package main

import (
	"flag"
	"log"
	"os"
	"sync"
)

var (
	goroutines        = flag.Int("workers", 100, "Initial number of goroutines to create during startup")
	pgcrStartingPoint = flag.Int64("pgcr", -1, "The starting point to fetch PGCRs from")
	passes            = flag.Int("passes", -1, "# of periods that crawler will do")
	period            = flag.Int64("period", 10000, "# of PGCRs the crawler will fetch in bursts")
	env               = flag.String("env", "local", "The environment that the CLI application should run on")
)

func prepareWorkers(wg *sync.WaitGroup, ids chan int64) {
	for i := 0; i < *goroutines; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, ids chan int64, id int) {
			defer wg.Done()
			for instanceId := range ids {
				log.Printf("Worker-%d: Processing instance Id [%d]", id, instanceId)
			}
		}(wg, ids, i)
	}
}

func main() {
	flag.Parse()
	run()
	os.Exit(0)
}
