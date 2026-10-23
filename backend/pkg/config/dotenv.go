package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads the first existing environment file from paths. Missing
// candidates are skipped, which is important on platforms such as Vercel,
// where configuration is injected directly and no .env file exists.
func LoadDotEnv(paths ...string) error {
	for _, path := range paths {
		if path == "" || !fileExists(path) {
			continue
		}
		if err := godotenv.Load(path); err != nil {
			return fmt.Errorf("load env file %s: %w", path, err)
		}
		return nil
	}
	return nil
}
