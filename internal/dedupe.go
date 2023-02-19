package internal

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func DedupeCmd(c *cli.Context) error {
	imagePaths, path, err := cmdBase(c)
	if err != nil {
		return err
	}

	slog := log.With().
		Str("cmd", c.Command.Name).
		Str("path", path).
		Logger()

	slog.Info().Int("imageCount", len(imagePaths)).
		Msg("Completed enumerating images.")

	// Generate perceptual hashes.
	slog.Info().Msg("Calculating hashes...")
	store := newHashStore()
	err = store.indexPaths(imagePaths)
	if err != nil {
		slog.Error().Err(err).Msg("Failed to calculate hashes.")
		return err
	}
	slog.Info().Msg("Completed calculating hashes.")

	// Query store for duplicates based on their hash.
	slog.Info().Msg("Finding duplicates...")
	duplicates := store.findDuplicates()
	slog.Info().Int("duplicateBatches", len(duplicates)).
		Msg("Completed finding duplicates.")

	// Copy or move the discovered duplicates.
	if c.Bool("copy") {
		slog.Info().Msg("Copying duplicates to subfolder...")
		err = handleDupes(path, duplicates, fileHandlerCopy)
		if err != nil {
			slog.Error().Err(err).Msg("Failed to copy duplicates.")
			return err
		}
		slog.Info().Msg("Completed copying duplicates to subfolder.")
	} else {
		slog.Info().Msg("Moving duplicates to subfolder...")
		err = handleDupes(path, duplicates, fileHandlerMove)
		if err != nil {
			slog.Error().Err(err).Msg("Failed to move duplicates.")
			return err
		}
		slog.Info().Msg("Completed moving duplicates to subfolder.")
	}
	return nil
}

func handleDupes(path string, duplicates [][]string, fn fileHandleFunc) error {
	dupePath := filepath.Join(path, "duplicates")

	if _, err := os.Stat(dupePath); os.IsNotExist(err) {
		err = os.Mkdir(dupePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for i, dupeset := range duplicates {
		prefix := fmt.Sprintf("set-%d-", i)
		for _, dupe := range dupeset {
			name := fmt.Sprintf("%s-%s", prefix, filepath.Base(dupe))
			err := fn(dupe, filepath.Join(dupePath, name))
			if err != nil {
				log.Warn().Err(err).
					Str("src", dupe).
					Msg("Failed to handle duplicate.")
			}
		}
	}

	return nil
}
