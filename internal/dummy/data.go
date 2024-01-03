package dummy

import (
	"gopkg.in/yaml.v3"
)

type Host struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type Gc struct {
	PoolSize int `yaml:"pool-size"`
}

type dummyConfigYaml struct {
	Host Host `yaml:"host"`
	Gc   Gc   `yaml:"gc"`
}

var DefaultConfigsBytes, _ = yaml.Marshal(&dummyConfigYaml{
	Host: Host{
		Name: "localhost",
		Port: 3000,
	},
	Gc: Gc{
		PoolSize: 4,
	},
})
