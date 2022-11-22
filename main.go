package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
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

func lastModifiedApp() *time.Time {
	filename := "main"
	// get last modified time
	file, err := os.Stat(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	modified := file.ModTime()
	return &modified
}
