package config

type Server struct {
	RunAddress      string
	DBConnectionURL string
}

func GetServerConfig() Server {
	return Server{
		RunAddress:      ":3200",
		DBConnectionURL: "",
	}
}
