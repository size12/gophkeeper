package config

// Server struct for server config.
type Server struct {
	RunAddress      string
	DBConnectionURL string
	FilesDirectory  string
}

// GetServerConfig gets server config.
func GetServerConfig() Server {
	return Server{
		RunAddress:      ":3200",
		DBConnectionURL: "",
		FilesDirectory:  "files",
	}
}
