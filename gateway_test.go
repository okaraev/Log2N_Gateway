package main

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func fakeSendMessagefail(iv interface{}) error {
	return fmt.Errorf("cannot send message")
}

func fakeSendMessageSuccess(iv interface{}) error {
	return nil
}

func TestValidate(t *testing.T) {
	myBson := bson.M{}
	err := Validate(myBson)
	if err == nil || fmt.Sprint(err) != "cannot find Team property or value" {
		t.Errorf("error must be 'cannot find Team property or value', instead of %v", err)
	}
	myBson["Team"] = "Team1"
	err = Validate(myBson)
	if err == nil || fmt.Sprint(err) != "cannot find Log property or value" {
		t.Errorf("error must be 'cannot find Log property or value', instead of %v", err)
	}
	myBson["Log"] = "Log"
	err = Validate(myBson)
	if err == nil || fmt.Sprint(err) != "cannot find Severity property or value" {
		t.Errorf("error must be 'cannot find Severity property or value', instead of %v", err)
	}
	myBson["Severity"] = "High"
	err = Validate(myBson)
	if err != nil {
		t.Errorf("error must be nil, instead of %v", err)
	}
}

func TestBreaker(t *testing.T) {
	FM := GetFMOverLoadInstace(fakeSendMessageSuccess)
	Brk := GetBreakerInstance(FM.MessageAdd)
	log := bson.M{}
	err := Brk.Do(log)
	if err != nil {
		t.Errorf("error must be nil, instead of %v", err)
	}
	FM.AddMessageFunction = fakeSendMessagefail
	Brk.Operation = FM.MessageAdd
	Brk.SuccessThreshold = 6 * time.Second
	Brk.OpenThreshold = 3 * time.Second
	for i := 1; i <= 4; i++ {
		err := Brk.Do(log)
		if err == nil {
			t.Errorf("error must not be nil")
		}
		if i == 4 {
			if fmt.Sprint(err) != "fail treshold exceeded" {
				t.Errorf("error message must be %s, instead of %s", "'fail treshold exceeded'", fmt.Sprint(err))
			}
			if Brk.Status != "Open" {
				t.Errorf("breaker status must be %s, instead of %s", "Open", Brk.Status)
			}
		} else {
			if Brk.FailCount != i {
				t.Errorf("failcount must be %d, instead of %d", i, Brk.FailCount)
			}
			if Brk.Status != "Closed" {
				t.Errorf("breaker status must be %s, instead of %s", "Closed", Brk.Status)
			}
		}
	}
	FM.AddMessageFunction = fakeSendMessageSuccess
	Brk.Operation = FM.MessageAdd
	time.Sleep(3 * time.Second)
	err = Brk.Do(log)
	if err != nil {
		t.Errorf("error must be nil instead of %v", err)
	}
	if Brk.Status != "HalfOpen" {
		t.Errorf("status must be %s instead of %s", "HalfOpen", Brk.Status)
	}
	time.Sleep(4 * time.Second)
	err = Brk.Do(log)
	if err != nil {
		t.Errorf("error must be nil instead of %v", err)
	}
	if Brk.FailCount != 0 {
		t.Errorf("FailCount must be %d instead of %d", 0, Brk.FailCount)
	}
	if Brk.Status != "Closed" {
		t.Errorf("breaker status must be %s, instead of %s", "Closed", Brk.Status)
	}
}
