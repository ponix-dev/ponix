package conf

type ManagementConfig struct {
	Port             string `env:"PORT"`
	DatabaseUrl      string `env:"DATABASE_URL"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`
	DatabaseUsername string `env:"DATABASE_USERNAME"`
}
