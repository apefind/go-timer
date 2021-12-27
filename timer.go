package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/gen2brain/beeep"
	"github.com/pterm/pterm"
)

func clearTerm() {
	print("\033[H\033[2J")
}

func captureKeyboardQuit(sig chan os.Signal) {
	for {
		key, _, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}
		if key == 'q' {
			sig <- os.Interrupt
		} else if key == 0 {
			sig <- syscall.SIGTERM
		}
	}
}

func beep(sig chan os.Signal) {
	for {
		select {
		case <-sig:
			return
		default:
			beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
			time.Sleep(1 * time.Second)
		}
	}
}

func Timer(duration time.Duration) {
	start := time.Now()
	stop := time.Now().Add(duration)
	clearTerm()
	pterm.Println(fmt.Sprintf("Start:              %s", start.Format(time.RFC850)))
	introSpinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).Start(fmt.Sprintf("Time remaining:            %s", duration))
	// introSpinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Time remaining:            %s", duration))
	sig := make(chan os.Signal)
	// signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go captureKeyboardQuit(sig)

	go func() {
		<-sig
		clearTerm()
		pterm.Println(fmt.Sprintf("Start:              %s", start.Format(time.RFC850)))
		pterm.Println(fmt.Sprintf("Stop:               %s", stop.Format(time.RFC850)))
		pterm.Println(fmt.Sprintf("Total time elapsed: %s", stop.Sub(start).Round(time.Second)))
		os.Exit(0)
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				introSpinner.Success(fmt.Sprintf("Time elapsed:       %s", duration))
				pterm.Println(fmt.Sprintf("Time elapsed:       %s", duration))
			case t := <-ticker.C:
				introSpinner.UpdateText(fmt.Sprintf("Time remaining:         %s", stop.Sub(t).Round(time.Second)))
			}
		}
	}()

	time.Sleep(duration)
	introSpinner.Stop()
	done <- true
	beep(sig)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <duration>\n", filepath.Base(os.Args[0]))
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
		flag.Usage()
		log.Println(err)
		os.Exit(1)
	}
	Timer(duration)
}
