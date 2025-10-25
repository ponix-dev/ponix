package conf

// AllInOne contains configuration for running all services in a single process.
// It embeds ManagementConfig and adds service-specific settings.
type AllInOne struct {
	Port string `env:"PORT"`
	ManagementConfig
}
