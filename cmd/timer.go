package main

import (
	"log"
	"os"

	"apefind/timer"
)

func main() {
	log.SetFlags(0)
	os.Exit(timer.TimerCmd(os.Args[1:]))
}
