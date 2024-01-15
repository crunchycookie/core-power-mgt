package model

type HostInfo struct {
	CPU            string   `json:"cpu"`
	SleepLevels    []string `json:"sleep-levels"`
	MaxAwakePower  float32  `json:"max-awake-power"`
	MaxAsleepPower float32  `json:"max-asleep-power"`
}

type SleepInfo struct {
	GcPoolSize int      `json:"gc-pool-size"`
	GcAsleep   int      `json:"gc-asleep"`
	GcAwake    int      `json:"gc-awake"`
	HostInfo   HostInfo `json:"host-info"`
}

type Host struct {
	Name      string `yaml:"name"`
	Port      int    `yaml:"port"`
	IsEmulate bool   `yaml:"is-emulate"`
}

type Topology struct {
	StableCoreCount  int `yaml:"stable-core-count"`
	DynamicCoreCount int `yaml:"dynamic-core-count"`
}

type PowerProfile struct {
	SleepIdleState string `yaml:"sleep-idle-state"`
	SleepFrq       int    `yaml:"sleep-frq"`
	PerfIdleState  string `yaml:"perf-idle-state"`
	PerfFrq        int    `yaml:"perf-frq"`
}

type ConfYaml struct {
	Host         Host         `yaml:"host"`
	Topology     Topology     `yaml:"topology"`
	PowerProfile PowerProfile `yaml:"power-profile"`
}

type GreenScore struct {
	AwakeStableCores  int `json:"avl-stable-cores"`
	UtilStableCores   int `json:"util-stable-cores"`
	AwakeDynamicCores int `json:"avl-dynamic-cores"`
	UtilDynamicCores  int `json:"util-dynamic-cores"`
	GreenScore        int `json:"green-score"`
}

type PowerStats struct {
	CpuType                string  `json:"cpu-type"`                        // ex: Intel xeon E3000
	HwUnitType             string  `json:"hw-unit-type"`                    // ex: cpu socket
	HwUnitPowerConsumption float32 `json:"hw-unit-power-consumption-watts"` // ex: 4.01 Watts
}
