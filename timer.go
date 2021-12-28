package main

import (
	"bufio"
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

type PTermWriter struct {
}

func (w *PTermWriter) Write(p []byte) (n int, err error) {
	pterm.Printo(string(p[:]))
	return len(p), nil
}

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

func Timer(duration time.Duration, interrupt chan os.Signal, w *bufio.Writer) error {
	msg := fmt.Sprintf("Timer %s: %%s", duration.Round(time.Second))
	_, err := w.WriteString(fmt.Sprintf(msg, duration.Round(time.Second)))
	if err != nil {
		return err
	}
	w.Flush()
	done := time.Now().Add(duration)
	beeping := false
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-interrupt:
			return nil
		case t := <-ticker.C:
			_, err := w.WriteString(fmt.Sprintf(msg, done.Sub(t).Round(time.Second)))
			if err != nil {
				return err
			}
			w.Flush()
			if !beeping && t.After(done) {
				go beep(interrupt)
				beeping = true
			}
		}
	}
}

func PTermTimer(duration time.Duration) {
	pterm.EnableColor()
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	w := &PTermWriter{}
	Timer(duration, interrupt, bufio.NewWriter(w))
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
