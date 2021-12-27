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

func keyboardInterrupt(sig chan os.Signal) {
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

func beep(interrupt chan os.Signal) {
	for {
		select {
		case <-interrupt:
			return
		default:
			beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
			time.Sleep(1 * time.Second)
		}
	}
}

func ptermOutput(output chan string) {
	pterm.EnableColor()
	for {
		select {
		case s := <-output:
			pterm.Printo(pterm.Red(s))
		}
	}
}

func Timer(duration time.Duration, interrupt chan os.Signal, out chan string) {
	msg := fmt.Sprintf("Timer %s: %%s ", duration.Round(time.Second))
	out <- fmt.Sprintf(msg, duration.Round(time.Second))
	done := time.Now().Add(duration)
	beeping := false
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-interrupt:
			return
		case t := <-ticker.C:
			out <- fmt.Sprintf(msg, done.Sub(t).Round(time.Second))
			if !beeping && t.After(done) {
				go beep(interrupt)
				beeping = true
			}
		}
	}
}

func PTermTimer(duration time.Duration) {
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	out := make(chan string)
	go ptermOutput(out)
	Timer(duration, interrupt, out)
	pterm.Println("")
	os.Exit(0)
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
	PTermTimer(duration)
}
