package power

import (
	"fmt"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"libvirt.org/go/libvirt"
	"slices"
)

func (o *SleepController) CalculateGreenScore(m *model.GreenScore) error {
	if o.isEmulate {
		return nil
	}
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
	m.UtilDynamicCores = int(utilDynamicCores)
	m.UtilStableCores = int(utilStableCores)

	utilMetric := utilDynamicCores + utilStableCores - uint(m.AwakeStableCores)
	if utilDynamicCores > 0 && utilMetric > 0 {
		m.GreenScore = int(utilMetric)
	} else {
		m.GreenScore = 0
	}

	return nil
}

func getCoreUtilizations(o *SleepController) (uint, uint, error) {

	//todo current version obtains utilization from libvirt virtualization. Need to make this configurable and extensible.
	return getCoreUtilizationFromLibvirt(o.sleepState.dynamicCpuIds, o.sleepState.stableCpuIds)
}

func getCoreUtilizationFromLibvirt(dynamicCoreIds []uint, stableCoreIds []uint) (uint, uint, error) {

	var pinnedInfo map[uint]bool
	for _, coreId := range dynamicCoreIds {
		pinnedInfo[coreId] = false
	}
	for _, coreId := range stableCoreIds {
		pinnedInfo[coreId] = false
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return -1, -1, fmt.Errorf("failed at connecting to the virtualization layer; libvirt: %w", err)
	}
	defer conn.Close()

	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return -1, -1, fmt.Errorf("failed at getting active instances information: %w", err)
	}

	fmt.Printf("%d running domains:\n", len(doms))
	for _, dom := range doms {
		name, err := dom.GetEmulatorPinInfo(1)
		if err != nil {
			return -1, -1, fmt.Errorf("failed at getting pinned info: %w", err)
		}
		if name != nil {
			pinnedInfo[uint(name)] = true
		}
		dom.Free()
	}

	utilStableCores := uint(0)
	utilDynamicCores := uint(0)
	for coreId, isPinned := range pinnedInfo {
		if !isPinned {
			continue
		}
		if slices.Contains(dynamicCoreIds, coreId) {
			utilDynamicCores++
		} else if slices.Contains(stableCoreIds, coreId) {
			utilStableCores++
		}
	}
	return utilDynamicCores, utilStableCores, nil
}

func (o *SleepController) ReadPowerStats(m *model.PowerStats) interface{} {
	if o.isEmulate {
		return nil
	}
}
