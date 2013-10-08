package main

import (
	"bufio"
	"io"
	"os"
	"github.com/funkygao/alser/parser"
	"sync"
	"time"
)

// Each single log file is a worker
func run_worker(logfile string, item jsonItem, wg *sync.WaitGroup, chLines chan int) {
	defer wg.Done()

	file, err := os.Open(logfile)
	if err != nil && err != os.ErrExist {
		// sometimes logs may be rotated, so ErrExist is common
		panic(err)
	}
	defer file.Close()

	if options.verbose {
		logger.Printf("%s started with %v", logfile, item.Parsers)
	}

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				if options.tailmode {
					time.Sleep(time.Second * time.Duration(tailSleep))
					continue
				} else {
					break
				}
			} else {
				panic(err)
			}
		}

		// a valid line scanned
		chLines <- 1
		for _, p := range item.Parsers {
			parser.Dispatch(p, string(line))
		}
	}

	logger.Println(logfile, "finished")
}
