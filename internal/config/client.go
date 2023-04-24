package config

type Client struct {
	ServerAddress string
}

func GetClientConfig() Client {
	return Client{
		ServerAddress: ":3200",
	}
}
