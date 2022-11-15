package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const (
	DDMMYYYY = "02.01.2006"
	AdminId  = "390579895"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var edu = newEdu()
	var telegramBot = newTelegram(edu)

	telegramBot.Run()
}

// KillExistsProcess убивает запущенный до этого процесс, чтобы не работало 2 демона одновременно
func KillExistsProcess() {
	b, err := os.ReadFile("pid.lock")
	if err == nil {
		var pid int
		pid, _ = strconv.Atoi(string(b))
		fmt.Printf("Kill process pid: %d \n", pid)
		proc := os.Process{Pid: pid}
		proc.Kill()
	}

	pid := os.Getpid()
	fmt.Printf("Save process pid: %d \n", pid)
	saveToFile("pid.lock", strconv.Itoa(pid))
}
