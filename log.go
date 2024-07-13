package main

import (
	"fmt"
	"os"
	"time"
)

const TIME_FORMAT = "2006-01-02 15:04:05"
const (
	C_RED_B    = "\033[91m"
	C_GREEN_B  = "\033[92m"
	C_YELLOW_B = "\033[93m"
	C_BLUE_B   = "\033[94m"
	C_CYAN_B   = "\033[96m"
	C_WHITE_B  = "\033[97m"
	C_RED      = "\033[31m"
	C_GREEN    = "\033[32m"
	C_YELLOW   = "\033[33m"
	C_BLUE     = "\033[34m"
	C_CYAN     = "\033[36m"
	C_WHITE    = "\033[37m"
)

var stepColors = []string{C_CYAN_B, C_CYAN, C_BLUE_B}

var logTimeStamp = false

func timeStamp(timeStamp time.Time) {
	if logTimeStamp {
		fmt.Print(C_WHITE + timeStamp.Format(TIME_FORMAT) + " ")
	}
}

func stepNum(stepNumber, totalSteps int) {
	fmt.Print(C_WHITE + fmt.Sprintf("[%d/%d] ", stepNumber, totalSteps))
}

func info(time time.Time, message string) {
	timeStamp(time)
	fmt.Println(C_WHITE_B + message)
}

func step(message string, stepNumber, totalSteps, depth int) {
	timeStamp(time.Now())

	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}

	stepNum(stepNumber, totalSteps)

	color := C_BLUE
	if depth < len(stepColors) {
		color = stepColors[depth]
	}

	fmt.Println(color + message)
}

func stepError(message string, stepNumber, totalSteps, depth int) {
	timeStamp(time.Now())

	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}

	stepNum(stepNumber, totalSteps)
	fmt.Println(C_RED_B + message)
}

func fatal(message string) {
	fmt.Println(C_RED_B + "[FATAL]: " + message)
	os.Exit(1)
}

func success(message string) {
	timeStamp(time.Now())
	fmt.Println(C_GREEN_B + message)
}
