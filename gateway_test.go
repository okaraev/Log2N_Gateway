package main

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

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
