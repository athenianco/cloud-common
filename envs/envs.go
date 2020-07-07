package envs

import "os"

// OneOfEnvs returns the first non-empty value from specified environment variables.
func OneOfEnvs(envs ...string) string {
	for _, env := range envs {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}
