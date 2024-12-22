package config

var GlobalConfig *RuntimeConfig

type RuntimeConfig struct {
	// Path to the OpenAPI spec file
	OpenAPISpecPath string
}

func InitConfig() {
	GlobalConfig = &RuntimeConfig{}
}
