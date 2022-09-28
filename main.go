package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DDMMYYYY = "02.01.2006"
)

const (
	QUARTER_1 = iota + 1
	QUARTER_2
	QUARTER_3
	QUARTER_4
)

func main() {
	var edu = newEdu()

	ReaderCommandStart(edu)

	// Run until CTRL+C.
	select {}
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
					ChildName: "Попов Кирилл Олегович",
					Date:      date,
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "day":
				ret := edu.getEduByDay(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
					Date:      date,
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "quarter":
				ret := edu.getEduByQuarter(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
					DiaryType: arguments[0],
				})
				for _, v := range ret {
					res2B, _ := json.Marshal(v)
					fmt.Println(string(res2B))
				}
			case "subject":
				ret := edu.getEduBySubject(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
					DiaryType: arguments[0],
					Subject:   strings.Join(arguments[1:], " "),
				})
				res2B, _ := json.Marshal(ret)
				fmt.Println(string(res2B))
			case "subjects":
				ret := edu.getSubjects(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
					DiaryType: arguments[0],
				})
				for _, v := range ret {
					fmt.Println(v)
				}
			}
		}
	}()
}
