package main

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
	"phoenix/minecraft"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	config := readConfig()
	address := config.Connection.RemoteAddress
	minecraft.Run(address)
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
}

func readConfig() config {
	c := config{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("error encoding default config: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatalf("error writing encoded default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	data, _ = toml.Marshal(c)
	if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
		log.Fatalf("error writing config file: %v", err)
	}
	return c
}
