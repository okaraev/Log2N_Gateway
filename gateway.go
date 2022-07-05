package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var GlobalConfig webconfig

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
	QConfig []qconfig
}

func throw(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvs() error {
	PQCSpath := os.Getenv("pqconnectionstringpath")
	PQName := os.Getenv("plogqname")
	PQServerAddress := os.Getenv("plogqserveraddress")
	SQCSpath := os.Getenv("sqconnectionstringpath")
	SQName := os.Getenv("slogqname")
	SQServerAddress := os.Getenv("slogqserveraddress")
	if PQCSpath == "" {
		return fmt.Errorf("cannot get environment variable pqconnectionstringpath")
	}
	if PQName == "" {
		return fmt.Errorf("cannot get environment variable plogqname")
	}
	if SQCSpath == "" {
		return fmt.Errorf("cannot get environment variable sqconnectionstringpath")
	}
	if SQName == "" {
		return fmt.Errorf("cannot get environment variable slogqname")
	}
	pqcsbytes, err := os.ReadFile(PQCSpath)
	if err != nil {
		return err
	}
	sqcsbytes, err := os.ReadFile(SQCSpath)
	if err != nil {
		return err
	}
	plogqpass := strings.Split(string(pqcsbytes), "\n")[0]
	slogqpass := strings.Split(string(sqcsbytes), "\n")[0]
	PQConnectionString := fmt.Sprintf("amqp://%s@%s", plogqpass, PQServerAddress)
	SQConnectionString := fmt.Sprintf("amqp://%s@%s", slogqpass, SQServerAddress)
	qconf1 := qconfig{QName: PQName, QConnectionString: PQConnectionString}
	qconf2 := qconfig{QName: SQName, QConnectionString: SQConnectionString}
	GlobalConfig.QConfig = []qconfig{qconf1, qconf2}
	return nil
}

func main() {
	err := getEnvs()
	throw(err)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/api/1/Log", AddLog)
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		throw(fmt.Errorf("cannot find http_port environment variable"))
	}
	router.Run(":" + port)
}
