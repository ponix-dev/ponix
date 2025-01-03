package conf

type AllInOne struct {
	Port string `env:"PORT"`
	ManagementConfig
}
