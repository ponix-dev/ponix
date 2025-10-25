package conf

// ManagementConfig contains configuration for the management service.
// It includes database connection settings and The Things Network (TTN) integration parameters.
type ManagementConfig struct {
	Port               string `env:"PORT"`
	DatabaseUrl        string `env:"DATABASE_URL"`
	DatabasePassword   string `env:"DATABASE_PASSWORD"`
	DatabaseUsername   string `env:"DATABASE_USERNAME"`
	Database           string `env:"DATABASE"`
	ApplicationId      string `env:"TTN_APPLICATION"`
	TTNApiKey          string `env:"TTN_API_KEY"`
	TTNApiCollaborator string `env:"TTN_API_COLLABORATOR"`
	TTNServerName      string `env:"TTN_SERVER_NAME"`
	TTNRegion          string `env:"TTN_REGION"`
}
