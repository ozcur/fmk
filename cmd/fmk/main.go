package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/ozcur/fmk/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const (
	usageText = "fmk [%s] [options] [path]"
)

// set by the linker
var gitHash string

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	dev := os.Getenv("DEV")
	if dev != "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		// Setup CPU profiling
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create cpu profile")
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal().Err(err).Msg("failed to start cpu profile")
		}
		defer pprof.StopCPUProfile()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	app := &cli.App{
		Name:      "fmk",
		Usage:     "a tool to quickly sort through your image datasets",
		Version:   gitHash,
		UsageText: fmt.Sprintf(usageText, "command"),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "copy",
				Usage: "copy files instead of moving them.",
			},
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:      "dedupe",
			Aliases:   []string{"d"},
			Usage:     "find duplicate images",
			Action:    internal.DedupeCmd,
			UsageText: fmt.Sprintf(usageText, "dedupe"),
		},
		{
			Name:      "sort",
			Aliases:   []string{"s"},
			Usage:     "sort images into directories",
			Action:    internal.SortCmd,
			UsageText: fmt.Sprintf(usageText, "sort"),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "folders",
					Aliases: []string{"f"},
					Value:   "keep,fix,trash",
					Usage:   "comma-separated list of folders to sort into",
				},
			},
		},
		{
			Name:      "rename",
			Aliases:   []string{"r"},
			Usage:     "rename images",
			Action:    internal.RenameCmd,
			UsageText: fmt.Sprintf(usageText, "rename"),
		},
		{
			Name:      "setsort",
			Aliases:   []string{"ss"},
			Usage:     "pick between images prefixed with set-*-*",
			Action:    internal.SetSortCmd,
			UsageText: fmt.Sprintf(usageText, "setsort"),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run app")
	}
}
