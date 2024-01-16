package power

import (
	"fmt"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"github.com/intel/power-optimization-library/pkg/power"
	"log"
	"slices"
	"sync"
)

const (
	StablePool                     = "stbl-pool"
	DynamicPool                    = "dyn-pool"
	MaxPerformancePowerProfileName = "maxPerfProf"
)

type CoreSleeps struct {
	stableCpuIds         []int
	dynamicCpuIds        []int
	isDynamicCoresAsleep bool
}

var DeepestSleepStateLbl string

type SleepController struct {
	Host       power.Host
	conf       model.ConfYaml
	mu         sync.Mutex
	isEmulate  bool
	sleepState CoreSleeps
}

func NewSleepController(conf *model.ConfYaml) (*SleepController, error) {

	if conf.Host.IsEmulate {
		log.Println("switching to emulation mode...")
		var stableCoreIds []uint
		for i := 0; i < conf.Topology.StableCoreCount; i++ {
			stableCoreIds = append(stableCoreIds, uint(i))
		}
		var dynamicCoreIds []uint
		for i := len(stableCoreIds); i < (len(stableCoreIds) + conf.Topology.DynamicCoreCount); i++ {
			dynamicCoreIds = append(dynamicCoreIds, uint(i))
		}
		return &SleepController{
			Host:       nil,
			conf:       *conf,
			isEmulate:  true,
			sleepState: getSleepState(stableCoreIds, dynamicCoreIds),
		}, nil
	}

	log.Println("creating a power instance...")
	host, err := getPowerHost()
	if err != nil {
		return nil, err
	}

	coreIds := host.GetAllCpus().IDs()
	reqCoreCount := conf.Topology.StableCoreCount + conf.Topology.DynamicCoreCount
	if reqCoreCount > len(coreIds) {
		return nil, fmt.Errorf("incorrect topology. required core count: %d, but only %d available", reqCoreCount, len(coreIds))
	}
	managedCoreIds := coreIds[0:reqCoreCount]
	stableCoreIds := managedCoreIds[0:conf.Topology.StableCoreCount]
	dynamicCoreIds := managedCoreIds[len(stableCoreIds):reqCoreCount]

	log.Printf("moving cores: %v into the shared pool...", managedCoreIds)
	err = host.GetSharedPool().SetCpuIDs(managedCoreIds)
	if err != nil {
		return nil, fmt.Errorf("failed at moving all cpu cores into the shared pool: %w", err)
	}

	log.Printf("grouping %v into stable and %v into dynamic pools...", stableCoreIds, dynamicCoreIds)
	err1 := moveCoresToPool(&host, StablePool, stableCoreIds)
	err2 := moveCoresToPool(&host, DynamicPool, dynamicCoreIds)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed at grouping cores into pools: %w and %w", err1, err2)
	}

	log.Println("setting initial perf levels: dynamic pool initially at sleep perf...")
	err1 = setPerf(&host, StablePool, uint(conf.PowerProfile.PerfFrq))
	err2 = setPerf(&host, DynamicPool, uint(conf.PowerProfile.SleepFrq))
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed at setting cores to max performance: %w and %w", err1, err2)
	}

	availableIdleStates := host.AvailableCStates()
	if !slices.Contains(availableIdleStates, conf.PowerProfile.PerfIdleState) || !slices.Contains(availableIdleStates, conf.PowerProfile.SleepIdleState) {
		return nil, fmt.Errorf("platform does not support requested idle states. need %s and %s, "+
			"but only supports %s", conf.PowerProfile.PerfIdleState, conf.PowerProfile.SleepIdleState, availableIdleStates)
	}

	log.Println("setting initial sleep levels: dynamic pool initially sleeps...")
	err1 = setPoolSleepState(&host, StablePool, conf.PowerProfile.PerfIdleState, availableIdleStates)
	err2 = setPoolSleepState(&host, DynamicPool, conf.PowerProfile.SleepIdleState, availableIdleStates)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed setting pool sleep states: %w, %w", err1, err2)
	}

	return &SleepController{
		Host:       host,
		conf:       *conf,
		isEmulate:  false,
		sleepState: getSleepState(stableCoreIds, dynamicCoreIds),
	}, nil
}

func getSleepState(stableCoreIds []uint, dynamicCoreIds []uint) CoreSleeps {
	var stableCpuIds []int
	for _, id := range stableCoreIds {
		stableCpuIds = append(stableCpuIds, int(id))
	}
	var dynamicCpuIds []int
	for _, id := range dynamicCoreIds {
		dynamicCpuIds = append(dynamicCpuIds, int(id))
	}
	sleepState := CoreSleeps{
		stableCpuIds:         stableCpuIds,
		dynamicCpuIds:        dynamicCpuIds,
		isDynamicCoresAsleep: true,
	}
	return sleepState
}

func (o *SleepController) Clean() error {
	if o.isEmulate {
		return nil
	}
	exlPools := o.Host.GetAllExclusivePools()
	err1 := exlPools.ByName(StablePool).Remove()
	err2 := exlPools.ByName(DynamicPool).Remove()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("failed at moving cores back to the shared pool: %w, %w", err1, err2)
	}
	err := o.Host.GetSharedPool().Remove()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("failed at moving cores back to the reserved pool: %w", err)
	}
	return nil
}

func setPoolSleepState(host *power.Host, poolName string, idleState string, avlIdleStates []string) error {
	cStates := power.CStates{}
	for _, state := range avlIdleStates {
		if state == idleState {
			cStates[state] = true
			continue
		}
		cStates[state] = false
	}
	log.Printf("setting pool: %s sleep levels as %v ...", poolName, cStates)
	err := (*host).GetExclusivePool(poolName).SetCStates(cStates)
	if err != nil {
		return fmt.Errorf("failed at setting %s pool to %v", poolName, cStates)
	}
	return nil
}

func setPerf(host *power.Host, poolName string, baseFMhz uint) error {
	maxFMhz := baseFMhz + 100
	maxPerfProf, err := power.NewPowerProfile(MaxPerformancePowerProfileName, baseFMhz, maxFMhz, "performance", "performance")
	if err != nil {
		return fmt.Errorf("failed at creating a power profile: %w", err)
	}
	err = (*host).GetExclusivePool(poolName).SetPowerProfile(maxPerfProf)
	if err != nil {
		return fmt.Errorf("failed at setting %s pool to max power profile: %w", poolName, err)
	}
	return nil
}

func getPowerHost() (power.Host, error) {
	host, allErrors := power.CreateInstance("gc-enabled-host")
	if host != nil {
		features := host.GetFeaturesInfo()
		var reqPowerOptmzFeatures error
		reqPowerOptmzFeatures = features[power.CStatesFeature].FeatureError()
		reqPowerOptmzFeatures = features[power.FreqencyScalingFeature].FeatureError()
		if reqPowerOptmzFeatures != nil {
			return nil, fmt.Errorf("failed at creating a power instance: %w", allErrors)
		}
	}
	return host, nil
}

func moveCoresToPool(host *power.Host, poolName string, coreIDs []uint) error {
	gcPool, err := (*host).AddExclusivePool(poolName)
	if err != nil {
		return fmt.Errorf("failed at creating exclusive pool for %s: %w", poolName, err)
	}
	err = gcPool.MoveCpuIDs(coreIDs)
	if err != nil {
		return fmt.Errorf("failed at moving cpu core to the %s pool: %w", poolName, err)
	}
	return nil
}
