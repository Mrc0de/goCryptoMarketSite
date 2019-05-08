package main

import (
	"log"
	"os"
	"time"
)

var (
	logger *log.Logger
)

func main() {
	logger = log.New(os.Stderr,"["+time.Now().Format(time.RFC850) + "] ",0)
	logger.Printf("Starting Services... \r\n")
	quit := make(chan string)
	go startWWWService(quit)
	byeString := <-quit
	logger.Println(byeString)
}