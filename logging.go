package main 

import (
	"fmt"
	"log"
	"runtime/debug"
)

func LogTwit(msg string) {
	log.Printf("twit %s\n", msg)
}

func LogInfo(msg string) {
	log.Printf("INFO %s\n", msg)
}

func LogWarning(msg string) {
	log.Printf("WARN %s\n", msg)
}

func LogError(msg string) {
	log.Printf("ERRO %s\n", msg)
	log.Println(string(debug.Stack()))
}

func LogCritical(msg string) {
	log.Printf("CRIT %s\n", msg)
	log.Println(string(debug.Stack()))
}

func LogTwitf(format string, a ...interface{}) {
	log.Printf("twit %s\n", fmt.Sprintf(format, a...))
}

func LogInfof(format string, a ...interface{}) {
	log.Printf("INFO %s\n", fmt.Sprintf(format, a...))
}

func LogWarningf(format string, a ...interface{}) {
	log.Printf("WARN %s\n", fmt.Sprintf(format, a...))
}

func LogErrorf(format string, a ...interface{}) {
	log.Printf("ERRO %s\n", fmt.Sprintf(format, a...))
	log.Println(string(debug.Stack()))
}

func LogCriticalf(format string, a ...interface{}) {
	log.Printf("CRIT %s\n", fmt.Sprintf(format, a...))
	log.Println(string(debug.Stack()))
}

