package main

import (
	"errors"
	"fmt"
	_ "t-tran/modules"
)

func main() {
	fmt.Println("input anykey to stop")
	var name string
	fmt.Scanln(&name)
}

// Account 账户
type Account struct {
	balance     float64
	deltaChan   chan float64
	balanceChan chan float64
	errChan     chan error
}

// NewAccount 创建账户
func NewAccount(balance float64) (a *Account) {
	a = &Account{
		balance:     balance,
		deltaChan:   make(chan float64),
		balanceChan: make(chan float64),
		errChan:     make(chan error),
	}
	go a.run()
	return
}

// Balance 返回账户余额
func (a *Account) Balance() float64 {
	return <-a.balanceChan
}

// Deposit 存入
func (a *Account) Deposit(amount float64) error {
	a.deltaChan <- amount
	return <-a.errChan
}

// Withdraw 撤销存入
func (a *Account) Withdraw(amount float64) error {
	a.deltaChan <- -amount
	return <-a.errChan
}

func (a *Account) applyDelta(amount float64) error {
	newBalance := a.balance + amount
	if newBalance < 0 {
		return errors.New("Insufficient funds")
	}
	a.balance = newBalance
	return nil
}

func (a *Account) run() {
	for {
		select {
		case delta := <-a.deltaChan:
			a.errChan <- a.applyDelta(delta)
		case a.balanceChan <- a.balance:
		}
	}
}
