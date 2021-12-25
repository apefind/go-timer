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
)

func clear() {
	print("\033[H\033[2J")
}

func Timer(duration time.Duration) error {
	start := time.Now()
	stop := time.Now().Add(duration)
	clear()
	msg := fmt.Sprintf("Time remaining: %s", duration)
	// introSpinner, _ := pterm.DefaultSpinner.WithShowTimer(false).WithRemoveWhenDone(true).Start(msg)
	introSpinner, _ := pterm.DefaultSpinner.Start(msg)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		msg := fmt.Sprintf("Start: %s", start.Format(time.RFC850))
		introSpinner.Success(msg)
		msg = fmt.Sprintf("Stop:  %s", time.Now().Format(time.RFC850))
		introSpinner.Success(msg)
		os.Exit(0)
	}()

	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				msg := fmt.Sprintf("Time elapsed: %s", duration)
				introSpinner.Success(msg)
				// if err := beeep.Notify(msg, "Time's up!", "assets/information.png"); err != nil {
				// 	return
				// }
				return
			case t := <-ticker.C:
				msg := fmt.Sprintf("Time remaining: %s", stop.Sub(t).Round(time.Second))
				// s, _ := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString(remaining)).Srender()
				introSpinner.UpdateText(msg)
			}
		}
	}()
	time.Sleep(duration)
	ticker.Stop()
	introSpinner.Stop()
	done <- true
	for {
		if err := beep.Beep(beep.DefaultFreq, beep.DefaultDuration); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	// return nil
}
func Usage(cmd string, flags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "\n%s %s -d <duration> command [args ...]\n\n", filepath.Base(os.Args[0]), cmd)
	fmt.Fprintf(os.Stderr, "\trun a command under time limitation\n\n")
}

func TimerCmd(args []string) int {
	var duration time.Duration
	flags := flag.NewFlagSet("timer", flag.ExitOnError)
	flags.DurationVar(&duration, "duration", 0*time.Second, "duration")
	flags.DurationVar(&duration, "d", 0*time.Second, "")
	flags.Usage = func() { Usage("", flags) }
	flags.Parse(args)
	log.Println(duration)
	if err := Timer(duration); err != nil {
		log.Println(err)
	}
	return 0
}

func main() {
	log.SetFlags(0)
	os.Exit(TimerCmd(os.Args[1:]))
}
