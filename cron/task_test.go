package cron

import (
	"fmt"
	"testing"
	"time"
)

var (
	testF = func() {
		time.Sleep(time.Microsecond * 1)

		nano := time.Now().UnixNano()
		if nano%2 == 0 {
			fmt.Println("panic")
			panic(1)
		} else {
			fmt.Println("running")
		}

	}
)

func TestNormalTask(t *testing.T) {
	task := newTask("", ModeNormal, testF)

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Microsecond * 50)

		go task.Run()
	}

	time.Sleep(time.Second * 2)
	if task.runNum != task.successTimes+task.panicTimes {
		fmt.Println("not equal")
	}
	fmt.Println(task.runNum, task.successTimes, task.panicTimes)
}

func TestWaitingTask(t *testing.T) {
	task := newTask("", ModeWaiting, testF)

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Microsecond * 50)

		go task.Run()
	}

	time.Sleep(time.Second * 3)
	if task.runNum != 1000 || task.runNum != task.successTimes+task.panicTimes {
		fmt.Println("not equal")
	}
	fmt.Println(task.runNum, task.successTimes, task.panicTimes)
}

func TestWaitingOneTask(t *testing.T) {
	task := newTask("", ModeWaitingOne, testF)

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Microsecond * 50)

		go task.Run()
	}

	time.Sleep(time.Second * 3)
	if task.runNum != task.successTimes+task.panicTimes {
		fmt.Println("not equal")
	}
	fmt.Println(task.runNum, task.successTimes, task.panicTimes)
}

func TestParallelTask(t *testing.T) {
	task := newTask("", ModeParallel, testF)

	for i := 0; i < 1000; i++ {
		time.Sleep(time.Microsecond * 50)

		go task.Run()
	}

	time.Sleep(time.Second * 3)
	if task.runNum != task.successTimes+task.panicTimes {
		fmt.Println("not equal")
	}
	fmt.Println(task.runNum, task.successTimes, task.panicTimes)
}
