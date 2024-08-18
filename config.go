package main

type (
	Config struct {
		ImportFilePath string   `env:"IMPORT_FILE_PATH"`
		OllamaModel    string   `env:"OLLAMA_MODEL, default=llama3.1"`
		HTTP           HTTP     `env:", prefix=HTTP_"`
		Postgres       Postgres `env:", prefix=POSTGRES_"`
	}

	Postgres struct {
		User     string `env:"USER, default=postgres"`
		Password string `env:"PASSWORD, default=postgres"`
		Database string `env:"DATABASE, default=postgres"`
		Host     string `env:"HOST, default=localhost"`
		Port     string `env:"PORT, default=5433"`
		SSLMode  string `env:"SSL_MODE, default=disable"`
	}

	HTTP struct {
		Host string `env:"HOST, default=0.0.0.0"`
		Port string `env:"PORT, default=8080"`
	}
)
