package internal

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type sortFlagError struct{}

func (e sortFlagError) Error() string {
	return "incorrect sort flag specified"
}

func SortCmd(c *cli.Context) error {
	imagePaths, path, err := cmdBase(c)
	if err != nil {
		return err
	}

	slog := log.With().
		Str("path", path).
		Str("cmd", c.Command.Name).
		Logger()

	// Sort out folders to.. sort to
	sortFlag := c.String("folders")
	sortFolders := strings.Split(sortFlag, ",")

	if len(sortFolders) < 2 {
		slog.Error().Str("folders", sortFlag).
			Msg("At least two folders must be specified.")
		return &sortFlagError{}
	}
	if len(sortFolders) > 9 {
		slog.Error().Str("folders", sortFlag).
			Msg("At most, nine folders can be specified.")
		return &sortFlagError{}
	}

	var sortPaths []string
	for _, folder := range sortFolders {
		sPath := filepath.Join(path, folder)
		if _, err := os.Stat(sPath); os.IsNotExist(err) {
			err = os.Mkdir(sPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
		sortPaths = append(sortPaths, sPath)
	}

	for i, sPath := range sortPaths {
		slog.Info().Str("folder", sPath).
			Int("hotkey", i+1).
			Str("path", sPath).
			Msg("Setup sorting folder.")
	}

	// Start GUI
	slog.Info().Msg("Starting GUI...")
	err = newSGUI(c, sortPaths, imagePaths).start()
	if err != nil {
		slog.Error().Err(err).Msg("GUI error.")
		return err
	}
	slog.Info().Msg("Complete.")

	return nil
}
