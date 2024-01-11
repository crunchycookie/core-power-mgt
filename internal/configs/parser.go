package configs

import (
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/dummy"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func NewConfigs(path string) *model.ConfYaml {

	var k = koanf.New(".")
	loadConfigs(path, k)
	return &model.ConfYaml{
		Host: model.Host{
			Name: k.String("host.name"),
			Port: k.Int("host.port"),
		},
		Topology: model.Topology{
			StableCoreCount:  k.Int("topology.stable-core-count"),
			DynamicCoreCount: k.Int("topology.dynamic-core-count"),
		},
		PowerProfile: model.PowerProfile{
			SleepIdleState: k.String("power-profile.sleep-idle-state"),
			SleepFrq:       k.Int("power-profile.sleep-frq"),
			PerfIdleState:  k.String("power-profile.perf-idle-state"),
			PerfFrq:        k.Int("power-profile.perf-frq"),
		},
	}
}

func loadConfigs(path string, k *koanf.Koanf) {
	var err error
	if len(path) > 0 {
		err = k.Load(file.Provider(path), yaml.Parser())
	} else {
		err = k.Load(rawbytes.Provider(dummy.DefaultConfigsBytes), yaml.Parser())
	}
	if err != nil {
		log.Fatalf("serviceerror loading config: %v", err)
	}
}
