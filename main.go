package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DDMMYYYY  = "02.01.2006"
	ChildName = "Попов Кирилл Олегович"
)

func main() {
	KillExistsProcess()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var edu = newEdu()
	var telegramBot = newTelegram(edu)

	telegramBot.Run()

	// Run until CTRL+C.
	select {}
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

func ReaderCommandStart(edu *Edu) {
	fmt.Println("Please write command:")

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			arguments := strings.Split(text, " ")
			text, arguments = arguments[0], arguments[1:]

			var date = time.Now().Format(DDMMYYYY)
			if len(arguments) > 0 {
				date = arguments[0]
			}

			switch text {
			case "login":
				edu.loginRequest("9046762614", arguments[0])
			case "week":
				ret := edu.getEduByWeek(&EduFilter{
					ChildName: ChildName,
					Date:      date,
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "day":
				ret := edu.getEduByDay(&EduFilter{
					ChildName: ChildName,
					Date:      date,
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "quarter":
				ret := edu.getEduByQuarter(&EduFilter{
					ChildName: ChildName,
					DiaryType: arguments[0],
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "subject":
				ret := edu.getEduBySubject(&EduFilter{
					ChildName: ChildName,
					DiaryType: arguments[0],
					Subject:   strings.Join(arguments[1:], " "),
				})
				res2B, _ := json.Marshal(ret)
				fmt.Println(string(res2B))
			case "subjects":
				ret := edu.getSubjects(&EduFilter{
					ChildName: ChildName,
					DiaryType: arguments[0],
				})
				for _, v := range ret {
					fmt.Println(v)
				}
			}
		}
	}()
}
