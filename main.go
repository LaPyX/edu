package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
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

const (
	ButtonMarks    = "marks"
	ButtonSchedule = "schedule"

	ButtonMarkDay     = "mark_day"
	ButtonMarkWeek    = "mark_week"
	ButtonMarkQuarter = "mark_quarter"
	ButtonMarkSubject = "mark_subject"

	ButtonScheduleToday    = "schedule_today"
	ButtonScheduleTomorrow = "schedule_tomorrow"
	ButtonScheduleWeek     = "schedule_week"
	ButtonScheduleDate     = "schedule_date"
)

var commandsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Оценки", ButtonMarks),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписание", ButtonSchedule),
	),
)

var markKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Оценки за день", ButtonMarkDay),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Оценки за неделю", ButtonMarkWeek),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Оценки за четверть", ButtonMarkQuarter),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Оценки по предмету", ButtonMarkSubject),
	),
)

var scheduleKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписание на сегодня", ButtonScheduleToday),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписание на завтра", ButtonScheduleTomorrow),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписание за неделю", ButtonScheduleWeek),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Расписание за дату", ButtonScheduleDate),
	),
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var edu = newEdu()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		//m, _ := json.Marshal(update)
		//log.Println("Message:" + string(m))

		if update.Message != nil {
			// Construct a new message from the given chat ID and containing
			// the text that we received.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")

			if update.Message.IsCommand() {
				// Extract the command from the Message.
				switch update.Message.Command() {
				case "start":
					msg.ReplyMarkup = commandsKeyboard
				case "close":
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				case "help":
					msg.Text = "I understand /sayhi and /status."
				default:
					msg.Text = "I don't know that command"
				}
			}

			// Send the message.
			if _, err = bot.Send(msg); err != nil {
				panic(err)
			}
		} else if update.CallbackQuery != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")

			switch update.CallbackQuery.Data {
			case ButtonMarks:
				msg.ReplyMarkup = markKeyboard
			case ButtonSchedule:
				msg.ReplyMarkup = scheduleKeyboard
			case ButtonMarkDay:
				ret := edu.getEduByDay(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
				})

				msg.Text = "Оценки за день: \n\n"
				var msgText []string
				for _, v := range ret {
					if v.Marks == nil {
						continue
					}

					var marks []string
					for _, m := range v.Marks {
						marks = append(marks, m.Value+"("+m.Reason+")")
					}
					msgText = append(msgText, v.Subject+": "+strings.Join(marks, ", "))
				}

				if msgText != nil {
					msg.Text += strings.Join(msgText, "\n")
				} else {
					msg.Text = "Оценок нет :("
				}
			case ButtonMarkWeek:
				ret := edu.getEduByWeek(&EduFilter{
					ChildName: "Попов Кирилл Олегович",
				})

				msg.Text = "Оценки за неделю: \n\n"
				var msgText []string
				for _, v := range ret {
					if v.Marks == nil {
						continue
					}

					var marks []string
					for _, m := range v.Marks {
						marks = append(marks, m.Value)
					}
					msgText = append(msgText, v.Subject+": "+strings.Join(marks, ", "))
				}

				if msgText != nil {
					msg.Text += strings.Join(msgText, "\n")
				} else {
					msg.Text = "Оценок нет :("
				}
			}

			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}

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
