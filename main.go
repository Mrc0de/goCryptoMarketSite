package main

import (
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	logger *log.Logger
	wwwConfig wwwServiceConfiguration
)

func main() {
	logger = log.New(os.Stderr,"["+time.Now().Format(time.RFC850) + "] ",0)
	logger.Println("Starting Services...")

	// Start Services
	quit := make(chan string)
	wwwConfig,err := loadConfig()
	if err != nil {
		// bail out
		logger.Printf("Cannot Start wwwService: %s", err)
		logger.Panic("Quitting.")
	}
	go startWWWService(quit,wwwConfig)

	// Setup Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	logger.Println("WWW Service Started.")

	// Select for channel signals on each active channel
	for  {
		select {
		case v := <- c:
			// Wait for Ctrl-C
			logger.Printf("[Control-C] Signal Received: %s",v)
			// Shut it all down
			// Tell WWWService to Shutdown
			quit <- "Shutdown"
			// Wait for Services to finish and return any string (which we will ignore)
			logger.Println("[WWW Service] " + <- quit )
			logger.Println("[goCryptoMarketSite] GoodBye.")
			os.Exit(0)
		default:
		}
	}
	logger.Println("-- Break --")
}