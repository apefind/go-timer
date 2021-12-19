package timer

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"

	"github.com/gen2brain/beeep"
	beep "github.com/gen2brain/beeep"
)

func clear() {
	print("\033[H\033[2J")
}

func Timer(duration time.Duration) error {
	clear()
	msg := fmt.Sprintf("Remaining time: %s", duration)
	introSpinner, _ := pterm.DefaultSpinner.WithShowTimer(false).WithRemoveWhenDone(true).Start(msg)
	stop := time.Now().Add(duration)
	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				if err := beeep.Notify(msg, "Time's up!", "assets/information.png"); err != nil {
					return
				}
				return
			case t := <-ticker.C:
				remaining := fmt.Sprintf("Remaining time: %s", stop.Sub(t).Round(time.Second))
				introSpinner.UpdateText(remaining)
			}
		}
	}()
	time.Sleep(duration)
	ticker.Stop()
	introSpinner.Stop()
	done <- true
	for i := 0; i < 20; i++ {
		if err := beep.Beep(beep.DefaultFreq, beep.DefaultDuration); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
