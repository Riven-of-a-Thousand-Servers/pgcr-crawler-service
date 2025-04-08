//go:build local

package main

import (
	"log"
	"sync"
	"time"
)

func run(goroutines int, period int64) {
	log.Printf(`Starting run with params:
      # of Workers: %d
      period: %d
      passes: %d
      starting point: %d`, goroutines, period, *passes, *pgcrStartingPoint)
	time.Sleep(5 * time.Second)

	var wg sync.WaitGroup

	for i := 1; i <= int(*passes); i++ {
		ids := make(chan int64, 50)
		prepareWorkers(&wg, ids)

		for j := 0; j < int(period); j++ {
			ids <- int64(j)
		}
		close(ids)
		wg.Wait()
		log.Printf("Finished pass [%d/%d]...\n", i, *passes)
	}

	log.Println("All workers have finished processing")
}
