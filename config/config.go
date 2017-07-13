package config

type Config struct {
	InitialSize int `json:"initial_size"`

	// A list of DHT neighbors.
	Neighbors []string `json:"neighbors"`
}
