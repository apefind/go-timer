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
	// color *pterm.Color
	style *pterm.Style
}

func NewPTermWriter(style *pterm.Style) *PTermWriter {
	return &PTermWriter{
		style: style,
	}
}

func (w *PTermWriter) Write(p []byte) (n int, err error) {
	if w.style != nil {
		pterm.Printo(w.style.Sprint(string(p[:])))
	} else {
		pterm.Printo(string(p[:]))
	}
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

func PTermTimer(duration time.Duration) error {
	pterm.EnableColor()
	defer pterm.Println("")
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	// style := pterm.NewStyle(pterm.FgRed, pterm.Bold)
	style := pterm.ThemeDefault.InfoMessageStyle
	w := NewPTermWriter(&style)
	// w := NewPTermWriter(nil)
	return Timer(duration, interrupt, bufio.NewWriter(w))
}

func StdoutTimer(duration time.Duration) error {
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	return Timer(duration, interrupt, bufio.NewWriter(os.Stdout))
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -o [pterm|stdout] <duration>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	var output string
	flag.StringVar(&output, "o", "pterm", "pterm or stdout")
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	duration, err := time.ParseDuration(flag.Args()[0])
	if err != nil {
		flag.Usage()
		log.Println(err)
		os.Exit(1)
	}
	if output == "pterm" {
		err = PTermTimer(duration)
	} else {
		err = StdoutTimer(duration)
	}
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
