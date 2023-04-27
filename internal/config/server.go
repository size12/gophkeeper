package config

type Server struct {
	RunAddress      string
	DBConnectionURL string
	FilesDirectory  string
}

func GetServerConfig() Server {
	return Server{
		RunAddress:      ":3200",
		DBConnectionURL: "",
		FilesDirectory:  "files",
	}
}
