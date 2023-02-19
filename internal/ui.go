package internal

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/theckman/yacspin"
)

var (
	spinCfg = yacspin.Config{
		Frequency:  50 * time.Millisecond,
		CharSet:    yacspin.CharSets[9],
		ShowCursor: true,
	}
)

type spinner struct {
	s *yacspin.Spinner
}

func newSpinner() *spinner {
	s, err := yacspin.New(spinCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create spinner")
	}

	err = s.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start spinner")
	}

	return &spinner{s: s}
}

func (s *spinner) Stop() {
	err := s.s.Stop()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to stop spinner")
	}
}

func (s *spinner) SetMessage(msg string) {
	s.s.Message(msg)
}
