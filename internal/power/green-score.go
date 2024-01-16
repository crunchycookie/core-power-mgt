package power

import (
	"fmt"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/utils"
	"slices"
	"strconv"
	"strings"
)

func (o *SleepController) CalculateGreenScore(m *model.GreenScore) error {
	m.AwakeStableCores = len(o.sleepState.stableCpuIds)
	if o.sleepState.isDynamicCoresAsleep {
		m.AwakeDynamicCores = 0
	} else {
		m.AwakeDynamicCores = len(o.sleepState.dynamicCpuIds)
	}

	utilDynamicCores, utilStableCores, err := getCoreUtilizations(o)
	if err != nil {
		return fmt.Errorf("failed at obtaining core utilization info: %w", err)
	}
	m.UtilDynamicCores = utilDynamicCores
	m.UtilStableCores = utilStableCores

	utilMetric := utilDynamicCores + utilStableCores - m.AwakeStableCores
	if utilDynamicCores > 0 && utilMetric > 0 {
		m.GreenScore = utilMetric
	} else {
		m.GreenScore = 0
	}

	return nil
}

func getCoreUtilizations(o *SleepController) (int, int, error) {

	//todo current version obtains utilization from libvirt virtualization. Need to make this configurable and extensible.
	return getCoreUtilizationFromLibvirt(o.sleepState.dynamicCpuIds, o.sleepState.stableCpuIds)
}

type domainsVirshModel struct {
	Name string `json:"Name"`
}
type emulatorPinVirshModel struct {
	EmulatorCPUAffinity string `json:"emulator: CPU Affinity"`
}

func getCoreUtilizationFromLibvirt(dynamicCoreIds []int, stableCoreIds []int) (int, int, error) {

	utilDynamicCores := 0
	utilStableCores := 0
	var domains []domainsVirshModel
	err := utils.RunThirdPartyClient[domainsVirshModel](&domains, "virsh-list-domains.sh")
	if err != nil {
		return -1, -1, err
	}
	for _, domain := range domains {
		var cpuAffinities []emulatorPinVirshModel
		err := utils.RunThirdPartyClient[emulatorPinVirshModel](&cpuAffinities, "virsh-domain-get-pinned-cpu-core.sh", domain.Name)
		if err != nil {
			return -1, -1, err
		}
		for _, cpuAffinity := range cpuAffinities {
			pinnedCore, _ := strconv.Atoi(strings.Split(cpuAffinity.EmulatorCPUAffinity, "*: ")[1])
			if slices.Contains(dynamicCoreIds, pinnedCore) {
				utilDynamicCores++
			} else if slices.Contains(stableCoreIds, pinnedCore) {
				utilStableCores++
			}
		}
	}
	return utilDynamicCores, utilStableCores, err
}
