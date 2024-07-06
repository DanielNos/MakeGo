package main

import (
	"fmt"
	"os"
	"time"
)

const TIME_FORMAT = "2006-01-02 15:04:05"

func info(timeStamp time.Time, message string) {
	fmt.Println("\033[37m" + timeStamp.Format(TIME_FORMAT) + " \033[97m" + message)
}

func logError(step int, message string) {
	fmt.Printf("\033[37m"+time.Now().Format(TIME_FORMAT)+" \033[97m[%d/%d] \033[91m%s\n", step, action-2, message)
}

func log(step int, message string) {
	fmt.Printf("\033[37m"+time.Now().Format(TIME_FORMAT)+" \033[97m[%d/%d] \033[96m%s\n", step, action-2, message)
}

func logStep(step, totalSteps int, message string) {
	fmt.Printf("\033[37m"+time.Now().Format(TIME_FORMAT)+"   \033[97m[%d/%d] \033[94m%s\n", step, totalSteps, message)
}

func logSubStep(step, totalSteps int, message string) {
	fmt.Printf("\033[37m"+time.Now().Format(TIME_FORMAT)+"     \033[97m[%d/%d] \033[94m%s\n", step, totalSteps, message)
}

func logStepError(step, totalSteps int, message string) {
	fmt.Printf("\033[37m"+time.Now().Format(TIME_FORMAT)+"   \033[97m[%d/%d] \033[91m[ERROR] %s\n", step, totalSteps, message)
}

func fatal(message string) {
	fmt.Println("\033[91m[FATAL]: " + message)
	os.Exit(1)
}

func success(message string) {
	fmt.Println("\033[37m" + time.Now().Format(TIME_FORMAT) + " \033[32m" + message)
}
