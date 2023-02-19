package internal

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Returns a slice of paths to all images in the discovered path,
// the discovered path itself, and possibly an error.
func cmdBase(c *cli.Context) ([]string, string, error) {
	slog := log.With().Str("cmd", c.Command.Name).Logger()

	path := getPath(c)
	if path == "." {
		slog.Info().Str("path", path).
			Msg("No path provided, using current directory.")
	}

	slog = slog.With().Str("path", path).Logger()

	// Find all image files.
	slog.Info().Msg("Enumerating images...")

	s := newSpinner()
	imagePaths, err := findImages(path)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enumerate images.")
	}
	s.Stop()

	if len(imagePaths) == 0 {
		log.Warn().Msg("No images found.")
		return nil, "", nil
	}

	log.Info().Int("imageCount", len(imagePaths)).
		Msg("Completed enumerating images.")

	return imagePaths, path, nil
}
