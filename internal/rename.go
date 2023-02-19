package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func RenameCmd(c *cli.Context) error {
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

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	for i, imagePath := range imagePaths {
		imgDir := filepath.Dir(imagePath)
		imgBase := filepath.Base(imagePath)
		imgExt := filepath.Ext(imgBase)

		imgExt = strings.ToLower(imgExt)
		if imgExt == ".jpeg" {
			imgExt = ".jpg"
		}

		newBase := fmt.Sprintf("%06d-%s", i, timestamp)
		newImgName := newBase + imgExt
		newImagePath := filepath.Join(imgDir, newImgName)

		err = os.Rename(imagePath, newImagePath)
		if err != nil {
			slog.Warn().Err(err).
				Str("oldPath", imagePath).
				Str("newPath", newImagePath).
				Msg("Failed to rename file.")
		}
	}

	slog.Info().Int("imageCount", len(imagePaths)).
		Msg("Completed renaming images.")

	return nil
}
