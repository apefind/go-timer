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
	msg := fmt.Sprintf("Timer %s: %%s", duration.Round(time.Second))
	pterm.EnableColor()
	pterm.Printo(fmt.Sprintf(msg, duration.Round(time.Second)))
	done := time.Now().Add(duration)
	sig := make(chan os.Signal)

	ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	go keyboardInterrupt(sig)

	beeping := false
	for {
		select {
		case <-sig:
			pterm.Println("")
			os.Exit(0)
		case t := <-ticker.C:
			pterm.Printo(fmt.Sprintf(msg, done.Sub(t).Round(time.Second)))
			if !beeping && t.After(done) {
				go beep(sig)
				beeping = true
			}
		}
	}
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
