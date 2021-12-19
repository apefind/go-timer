package timer

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

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
