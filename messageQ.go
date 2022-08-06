package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/streadway/amqp"
)

type Breaker struct {
	Status           string        // Breaker Status
	FailCount        int           // Failed operations' count
	LastFail         time.Time     // Succeeded operations' count
	FailThreshold    int           // Failed operations threshold
	SuccessThreshold time.Duration // Time duration in which all operations must be succeeded after that FailCount will reset and Status will change to 'Closed'
	OpenThreshold    time.Duration // Time duration after which Status will change to 'HalfOpen'
	Operation        func(i interface{}) error
}

type FileManager struct {
	AddMessageFunction func(message interface{}) error
}

func (f FileManager) MessageAdd(message interface{}) error {
	result := f.AddMessageFunction(message)
	return result
}

func MessageSend(message interface{}) error {
	connectRabbitMQ, err := amqp.Dial(QConnectionString)
	if err != nil {
		return err
	}
	defer connectRabbitMQ.Close()
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channelRabbitMQ.Close()
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	Message := amqp.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	}
	err = channelRabbitMQ.Publish("", QName, false, false, Message)
	if err != nil {
		return err
	}
	return nil
}

func GetFMDefaultInstance() FileManager {
	FM := FileManager{AddMessageFunction: MessageSend}
	return FM
}

func GetFMOverLoadInstace(function func(message interface{}) error) FileManager {
	fm := FileManager{AddMessageFunction: function}
	return fm
}

func GetBreakerInstance(function func(iv interface{}) error) Breaker {
	br := Breaker{
		Status:           "Closed",
		OpenThreshold:    30 * time.Second,
		FailCount:        0,
		FailThreshold:    3,
		LastFail:         time.Now(),
		SuccessThreshold: 1 * time.Minute,
		Operation:        function,
	}
	return br
}

func (b *Breaker) Open() {
	b.Status = "Open"
	go func() {
		time.Sleep(b.OpenThreshold)
		b.Status = "HalfOpen"
	}()
}

func (b *Breaker) Do(iv interface{}) error {
	// IF Connection is OK and Fail threshold is exceeded mark connection as fail for a time for a fast fail
	if b.Status == "Closed" && b.FailCount >= b.FailThreshold {
		b.Open()
		return errors.New("fail treshold exceeded")
	}
	// IF connection marked as fail, return immediate error
	if b.Status == "Open" {
		return errors.New("fail treshold exceeded")
	}
	// DO operation and check result
	err := b.Operation(iv)
	if err != nil {
		if b.Status == "HalfOpen" {
			b.Open()
		} else {
			b.FailCount++
		}
		b.LastFail = time.Now()
		return err
	}
	// IF connection is marked as Healthy or halfHealthy check for Last Failed time
	if b.Status == "HalfOpen" || b.Status == "Closed" {
		if time.Since(b.LastFail) >= b.SuccessThreshold {
			b.Close()
		}
	}
	return nil
}

func (b *Breaker) Close() {
	b.Status = "Closed"
	b.FailCount = 0
}
