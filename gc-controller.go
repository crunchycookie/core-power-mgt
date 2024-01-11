package main

import (
	"fmt"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/configs"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/handler"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/power"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/serviceerror"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"strconv"
)

func main() {

	log.Println("loading service configurations...")
	conf := loadConfigs()

	log.Println("creating an api handler...")
	apiHandler, err := getAPIHandler(conf)
	if err != nil {
		fmt.Println("Failed to create an API handler", err)
		return
	}
	attachCleanUponShutdownHandler(apiHandler)

	log.Println("configuring api routing...")
	router := gin.Default()
	router.Use(serviceerror.ErrorHandler())

	router.GET("/gc-controller/sleep-info", apiHandler.GetSleepInfo)
	router.PUT("/gc-controller/sleep", apiHandler.PutSleepOP)
	router.PUT("/gc-controller/wake", apiHandler.PutAwakeOP)
	router.PUT("/gc-controller/dev/perf", apiHandler.PutPoolFreq)

	log.Println("begin serving...")
	err = router.Run(conf.Host.Name + ":" + strconv.Itoa(conf.Host.Port))
	if err != nil {
		fmt.Println("Unable to start the gc-controller", err)
		return
	}
}

func attachCleanUponShutdownHandler(apiHandler *handler.SleepAPIHandler) {
	// Set cleanup.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("[service quit signal received]")
		cleanup(apiHandler)
		os.Exit(0)
	}()
}

func cleanup(handler *handler.SleepAPIHandler) {
	log.Println("restoring core power management...")
	err := handler.Clean()
	if err != nil {
		log.Fatal("unable to restore the system. there can be residue of core power management that might "+
			"degrade hardware. consider rebooting to properly hand-over power management to the os. ", err)
	}
	log.Println("power management safely handed over to the os. Goodbye!")
}

func getAPIHandler(conf *model.ConfYaml) (*handler.SleepAPIHandler, error) {
	controller, err := power.NewSleepController(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sleep controller: %w", err)
	}
	sleepHandler := handler.SleepAPIHandler{
		Controller: controller,
	}
	return &sleepHandler, nil
}

func loadConfigs() *model.ConfYaml {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	conf := configs.NewConfigs(path)
	return conf
}
