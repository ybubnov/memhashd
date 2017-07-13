package config

type Config struct {
	// A list of DHT neighbors.
	Neighbors []string `json:"neighbors"`
}
