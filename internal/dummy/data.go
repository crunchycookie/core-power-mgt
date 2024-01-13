package dummy

import (
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"gopkg.in/yaml.v3"
)

var DefaultConfigsBytes, _ = yaml.Marshal(&model.ConfYaml{
	Host: model.Host{
		Name:      "localhost",
		Port:      3000,
		IsEmulate: false,
	},
	Topology: model.Topology{
		StableCoreCount:  4,
		DynamicCoreCount: 1,
	},
	PowerProfile: model.PowerProfile{
		SleepIdleState: "C3_ACPI",
		SleepFrq:       400,
		PerfIdleState:  "POLL",
		PerfFrq:        2800,
	},
})
