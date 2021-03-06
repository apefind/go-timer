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
	beep "github.com/gen2brain/beeep"
	"github.com/pterm/pterm"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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

func Timer(duration time.Duration, interrupt chan os.Signal, w *bufio.Writer) error {
	msg := fmt.Sprintf("Timer %s: %%s", duration.Round(time.Second))
	if _, err := w.WriteString(fmt.Sprintf(msg, duration.Round(time.Second))); err != nil {
		return err
	}
	w.Flush()
	done := time.Now().Add(duration)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-interrupt:
			return nil
		case t := <-ticker.C:
			if _, err := w.WriteString(fmt.Sprintf(msg, done.Sub(t).Round(time.Second))); err != nil {
				return err
			}
			w.Flush()
			if t.After(done) {
				beep.Beep(beep.DefaultFreq, beep.DefaultDuration)
			}
		}
	}
}

type FileWriter struct {
	*os.File
}

func NewFileWriter(w *os.File) *FileWriter {
	return &FileWriter{w}
}

func (w FileWriter) Write(p []byte) (int, error) {
	return w.WriteString(string(p[:]) + "\n")
}

func StdoutTimer(duration time.Duration) error {
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	return Timer(duration, interrupt, bufio.NewWriter(NewFileWriter(os.Stdout)))
}

type PTermWriter struct {
	area  *pterm.AreaPrinter
	style *pterm.Style
}

func NewPTermWriter(area *pterm.AreaPrinter, style *pterm.Style) *PTermWriter {
	return &PTermWriter{
		area:  area,
		style: style,
	}
}

func (w *PTermWriter) Write(p []byte) (n int, err error) {
	if w.style != nil {
		w.area.Update(w.style.Sprint(string(p[:])))
	} else {
		w.area.Update(string(p[:]))
	}
	return len(p), nil
}

func PTermTimer(duration time.Duration, style *pterm.Style) error {
	pterm.EnableColor()
	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return err
	}
	defer area.Stop()
	interrupt := make(chan os.Signal)
	go keyboardInterrupt(interrupt)
	return Timer(duration, interrupt, bufio.NewWriter(NewPTermWriter(area, style)))
}

type FyneWriter struct {
	label *widget.Label
}

func NewFyneWriter(label *widget.Label) *FyneWriter {
	return &FyneWriter{
		label: label,
	}
}

func (w *FyneWriter) Write(p []byte) (n int, err error) {
	w.label.SetText(string(p[:]))
	return len(p), nil
}

func FyneTimer(duration time.Duration, style *pterm.Style) error {
	interrupt := make(chan os.Signal)
	app := app.New()
	win := app.NewWindow("Timer")
	label := widget.NewLabel("")
	quit := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { interrupt <- os.Interrupt; app.Quit() })
	grid := container.New(layout.NewHBoxLayout(), label, quit)
	win.SetContent(grid)
	win.Show()
	go Timer(duration, interrupt, bufio.NewWriter(NewFyneWriter(label)))
	app.Run()
	return nil
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -o [pterm|stdout] <duration>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	var output, style string
	flag.StringVar(&output, "o", "pterm", "output = fyne, pterm or stdout")
	flag.StringVar(&style, "s", "", "pterm primary style")
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
		if style == "" {
			err = PTermTimer(duration, nil)
		} else {
			err = PTermTimer(duration, &pterm.ThemeDefault.PrimaryStyle)
		}

	} else if output == "fyne" {
		err = FyneTimer(duration, nil)
	} else {
		err = StdoutTimer(duration)
	}
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
