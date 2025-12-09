package config

// GetMusicDir returns the configured music directory
func (c *Config) GetMusicDir() string {
	return c.System.MusicDir
}
