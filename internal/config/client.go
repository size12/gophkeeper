package config

// Client struct for client config.
type Client struct {
	ServerAddress string
}

// GetClientConfig gets client config.
func GetClientConfig() Client {
	return Client{
		ServerAddress: ":3200",
	}
}
