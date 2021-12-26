package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	beep "github.com/gen2brain/beeep"
	"github.com/pterm/pterm"

	"github.com/eiannone/keyboard"
)

func clearTerm() {
	print("\033[H\033[2J")
}

func Timer(duration time.Duration) {
	start := time.Now()
	stop := time.Now().Add(duration)
	clearTerm()
	pterm.Info.Println(fmt.Sprintf("Start:              %s", start.Format(time.RFC850)))
	introSpinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).Start(fmt.Sprintf("Time remaining:            %s", duration))
	// introSpinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Time remaining:            %s", duration))
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			char, _, err := keyboard.GetSingleKey()
			if err != nil {
				panic(err)
			}
			if char == 'q' {
				sig <- os.Interrupt
			}
		}
	}()

	go func() {
		<-sig
		stop := time.Now()
		clearTerm()
		pterm.Info.Println(fmt.Sprintf("Start:              %s", start.Format(time.RFC850)))
		pterm.Info.Println(fmt.Sprintf("Stop:               %s", stop.Format(time.RFC850)))
		pterm.Info.Println(fmt.Sprintf("Total time elapsed: %s", stop.Sub(start).Round(time.Second)))
		os.Exit(0)
	}()

	ticker := time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				introSpinner.Success(fmt.Sprintf("Time elapsed:       %s", duration))
				pterm.Info.Println(fmt.Sprintf("Time elapsed:       %s", duration))
				// return
			case t := <-ticker.C:
				introSpinner.UpdateText(fmt.Sprintf("Time remaining:         %s", stop.Sub(t).Round(time.Second)))
			}
		}
	}()

	time.Sleep(duration)
	ticker.Stop()
	introSpinner.Stop()
	done <- true

	for {
		beep.Beep(beep.DefaultFreq, beep.DefaultDuration)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s <duration>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	log.SetFlags(0)
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	duration, err := time.ParseDuration(os.Args[1])
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	Timer(duration)
}
