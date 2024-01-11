package handler

import (
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/model"
	"github.com/crunchycookie/openstack-gc/gc-controller/internal/power"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SleepAPIHandler struct {
	Controller *power.SleepController
}

func (o *SleepAPIHandler) GetSleepInfo(c *gin.Context) {
	postBody := o.Controller.Info()
	c.IndentedJSON(http.StatusOK, postBody)
}

func (o *SleepAPIHandler) PutSleepOP(c *gin.Context) {
	controller := o.Controller
	err := controller.Sleep()
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusCreated, nil)
}

func (o *SleepAPIHandler) PutAwakeOP(c *gin.Context) {
	controller := o.Controller
	err := controller.Wake()
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusCreated, nil)
}

func (o *SleepAPIHandler) PutPoolFreq(c *gin.Context) {
	var newFqOp model.FqOp
	if err := c.BindJSON(&newFqOp); err != nil {
		return
	}

	controller := o.Controller
	err := controller.OpFrequency(newFqOp.FMhz)
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newFqOp)
}

func (o *SleepAPIHandler) Clean() error {
	return o.Controller.Clean()
}
