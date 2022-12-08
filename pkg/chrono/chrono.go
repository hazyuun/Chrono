package chrono

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const DotChronoDirName string = ".chrono"
const SessionsFileName string = "sessions.json"

var RootPath string

func Init(path string) {
	RootPath = path

	cp := filepath.Join(path, DotChronoDirName)
	err := os.MkdirAll(cp, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create .chrono directory")
	}

	sp := filepath.Join(path, DotChronoDirName, SessionsFileName)

	_, err = os.ReadFile(sp)
	if os.IsNotExist(err) {
		f, err := os.Create(sp)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create sessions file")
		}

		_, err = f.WriteString("{}")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to write to sessions file")
		}

		f.Close()
	}
}
