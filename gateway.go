package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var GlobalConfig webconfig
var Brkr Breaker
var FM FileManager
var QConnectionString string
var QName string

type Log struct {
	Team     string
	Severity string
	Log      string
}

type qconfig struct {
	QConnectionString string
	QName             string
}

type webconfig struct {
	QConfig   []qconfig
	HTTP_PORT string
}

func throw(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvs() error {
	myQConfig := []qconfig{}
	for _, each := range []string{"p", "s"} {
		for _, item := range []string{"logqname", "logqserveraddress", "qconnectionstringpath"} {
			envVar := fmt.Sprintf("%s%s", each, item)
			log.Printf("Getting value of: %s\n", envVar)
			varValue := os.Getenv(envVar)
			if varValue == "" {
				return fmt.Errorf("cannot get environment variable %s", envVar)
			}
		}
		QCSpath := os.Getenv(fmt.Sprintf("%sqconnectionstringpath", each))
		QName := os.Getenv(fmt.Sprintf("%slogqname", each))
		QServerAddress := os.Getenv(fmt.Sprintf("%slogqserveraddress", each))
		qcsbytes, err := os.ReadFile(QCSpath)
		if err != nil {
			return err
		}
		logqpass := strings.Split(string(qcsbytes), "\n")[0]
		QConnectionString := fmt.Sprintf("amqp://%s@%s", logqpass, QServerAddress)
		qconf := qconfig{QName: QName, QConnectionString: QConnectionString}
		myQConfig = append(myQConfig, qconf)
	}
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		return fmt.Errorf("cannot find http_port environment variable")
	}
	GlobalConfig.HTTP_PORT = port
	GlobalConfig.QConfig = myQConfig
	return nil
}

func main() {
	err := getEnvs()
	throw(err)
	FM = GetFMDefaultInstance()
	Brkr = GetBreakerInstance(FM.MessageAdd)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/api/1/Log", AddLog)
	router.Run(":" + GlobalConfig.HTTP_PORT)
}
