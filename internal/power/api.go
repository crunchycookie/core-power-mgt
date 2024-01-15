package power

import (
	"fmt"
	"log"
)

func (o *SleepController) Info() map[string]string {
	fmt.Println("Listing out available sleep states...")
	if o.isEmulate {
		return map[string]string{
			"message": "Controller is in emulation mode. All api responses are replied with the happy path response",
		}
	}
	log.Printf("avl sleep states: %v\n", o.Host.AvailableCStates())
	return map[string]string{
		"avl-idle-states":  fmt.Sprintf("%v", o.Host.AvailableCStates()),
		"sleep-idle-state": o.conf.PowerProfile.SleepIdleState,
		"perf-idle-state":  o.conf.PowerProfile.PerfIdleState,
		"perf-fq":          fmt.Sprintf("%d", o.conf.PowerProfile.PerfFrq),
		"sleep-fq":         fmt.Sprintf("%d", o.conf.PowerProfile.SleepFrq),
	}
}

func (o *SleepController) Sleep() error {
	if o.isEmulate {
		return nil
	}

	(*o).mu.Lock()
	defer (*o).mu.Unlock()

	host := o.Host
	//todo need to support per-core sleep state and set it according to the coreCount parameter.
	// below only set pool sleep state.
	err2 := setPerf(&host, DynamicPool, uint(o.conf.PowerProfile.SleepFrq))
	err1 := setPoolSleepState(&host, DynamicPool, o.conf.PowerProfile.SleepIdleState, o.Host.AvailableCStates())
	if err1 != nil || err2 != nil {
		//todo handle serviceerror from calling level, and then we can remove below log.
		return fmt.Errorf("failed at sleeping dynamic pool: %w, %w", err1, err2)
	}
	log.Printf("dynamic pool sleep state changed to: %s", DeepestSleepStateLbl)

	// dynamic cores went to sleep.
	o.sleepState.isDynamicCoresAsleep = true
	return nil
}

func (o *SleepController) Wake() error {
	if o.isEmulate {
		return nil
	}

	(*o).mu.Lock()
	defer (*o).mu.Unlock()

	host := o.Host
	//todo need to support per-core sleep state and set it according to the coreCount parameter.
	// below only set pool sleep state.
	err2 := setPerf(&host, DynamicPool, uint(o.conf.PowerProfile.PerfFrq))
	err1 := setPoolSleepState(&host, DynamicPool, o.conf.PowerProfile.PerfIdleState, o.Host.AvailableCStates())
	if err1 != nil || err2 != nil {
		//todo handle serviceerror from calling level, and then we can remove below log.
		return fmt.Errorf("failed at waking dynamic pool: %w, %w", err1, err2)
	}
	log.Println("dynamic pool woken up")
	o.sleepState.isDynamicCoresAsleep = false
	return nil
}

func (o *SleepController) OpFrequency(fMhz uint) error {
	if o.isEmulate {
		return nil
	}

	(*o).mu.Lock()
	defer (*o).mu.Unlock()

	host := o.Host
	err := setPerf(&host, DynamicPool, fMhz)
	if err != nil {
		//todo handle serviceerror from calling level, and then we can remove below log.
		log.Print("failed at changing perf frequency: %w", err)
		return fmt.Errorf("failed at changing perf frequency: %w", err)
	}
	log.Printf("frequency of pool: %s changed to: %d", DynamicPool, fMhz)
	return nil
}
