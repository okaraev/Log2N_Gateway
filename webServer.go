package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type httpresponse struct {
	Status  bool
	Message string
}

func Validate(log bson.M) error {
	if val, ok := log["Team"]; !ok || val == "" {
		return fmt.Errorf("cannot find Team property or value")
	}
	if val, ok := log["Log"]; !ok || val == "" {
		return fmt.Errorf("cannot find Log property or value")
	}
	if val, ok := log["Severity"]; !ok || val == "" {
		return fmt.Errorf("cannot find Severity property or value")
	}
	return nil
}

func AddMessage(log bson.M) error {
	myBreaker := Breaker{}
	myBreaker.New(SendMessage)
	err := myBreaker.Do(GlobalConfig.QConfig[0].QConnectionString, GlobalConfig.QConfig[0].QName, log)
	if err != nil {
		err := SendMessage(GlobalConfig.QConfig[1].QConnectionString, GlobalConfig.QConfig[1].QName, log)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddLog(c *gin.Context) {
	message := "TestResponse"
	log := bson.M{}
	err := c.BindJSON(&log)
	if err != nil {
		message = fmt.Sprintf("Error: %s", err)
		c.IndentedJSON(400, httpresponse{Status: false, Message: message})
		return
	}
	err = Validate(log)
	if err != nil {
		message = fmt.Sprintf("Error: %s", err)
		c.IndentedJSON(400, httpresponse{Status: false, Message: message})
		return
	}
	err = AddMessage(log)
	if err != nil {
		message = fmt.Sprintf("Error: %s", err)
		c.IndentedJSON(400, httpresponse{Status: false, Message: message})
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: message})
}
