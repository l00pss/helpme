package haconfig

import (
	"fmt"
	"os"
)

type SimpleConfig struct {
	Server ServerConfig
}

func main() {
	// Set environment variables
	os.Setenv("SERVER_HOST", "test-host")
	os.Setenv("SERVER_PORT", "9000")

	config := New()
	var cfg SimpleConfig

	err := config.Load(&cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Config: %+v\n", cfg)
	fmt.Printf("Server: %+v\n", cfg.Server)
}
