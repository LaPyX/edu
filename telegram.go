package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Quarter1 = "1"
	Quarter2 = "2"
	Quarter3 = "3"
	Quarter4 = "4"
)

const (
	ButtonStart  = "start"
	ButtonFinish = "finish"

	ButtonMarks    = "marks"
	ButtonSchedule = "schedule"

	ButtonMarkDay           = "mark_day"
	ButtonMarkWeek          = "mark_week"
	ButtonMarkQuarter       = "mark_quarter"
	ButtonMarkSubject       = "mark_subject"
	ButtonSelectMarkSubject = "select_mark_subject"
	ButtonSelectMarkQuarter = "select_mark_quarter"

	ButtonScheduleToday      = "schedule_today"
	ButtonScheduleTomorrow   = "schedule_tomorrow"
	ButtonScheduleWeek       = "schedule_week"
	ButtonScheduleDate       = "schedule_date"
	ButtonSelectScheduleDate = "select_schedule_date"
)

var ListQuarter = []string{Quarter1, Quarter2, Quarter3, Quarter4}

var commandsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки", ButtonMarks),
	InlineKeyboardButtonRow("Расписание", ButtonSchedule),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
)

var markKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки за день", ButtonMarkDay),
	InlineKeyboardButtonRow("Оценки за неделю", ButtonMarkWeek),
	InlineKeyboardButtonRow("Оценки за четверть", ButtonMarkQuarter),
	InlineKeyboardButtonRow("Оценки по предмету", ButtonMarkSubject),
	BackKeyboardButtonRow(ButtonStart),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
)

var scheduleKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Расписание на сегодня", ButtonScheduleToday),
	InlineKeyboardButtonRow("Расписание на завтра", ButtonScheduleTomorrow),
	InlineKeyboardButtonRow("Расписание за неделю", ButtonScheduleWeek),
	InlineKeyboardButtonRow("Расписание за дату", ButtonScheduleDate),
	BackKeyboardButtonRow(ButtonStart),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
)

type Telegram struct {
	bot *tgbotapi.BotAPI
	edu *Edu
}

func newTelegram(edu *Edu) *Telegram {
	return &Telegram{
		edu: edu,
	}
}

func (t *Telegram) Run() {
	var err error

	t.bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	t.bot.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	log.Printf("Authorized on account %s", t.bot.Self.UserName)

	t.GetUpdatesChan()
}

func (t *Telegram) GetUpdatesChan() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		//m, _ := json.Marshal(update)
		//fmt.Println("---------------------")
		//fmt.Println(string(m))
		//fmt.Println("---------------------")

		if update.Message != nil {
			// Construct a new message from the given chat ID and containing
			// the text that we received.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")

			if update.Message.IsCommand() {
				// Extract the command from the Message.
				switch update.Message.Command() {
				case ButtonStart:
					msg.ReplyMarkup = commandsKeyboard
				}
			}

			// Send the message.
			if _, err := t.bot.Send(msg); err != nil {
				panic(err)
			}
		} else if update.CallbackQuery != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Выберите действие:")
			t.RemoveCallbackMessage(update.CallbackQuery)
			data, arguments := t.getData(update.CallbackQuery.Data)

			switch data {
			case ButtonStart:
				msg.ReplyMarkup = commandsKeyboard
			case ButtonMarks:
				msg.ReplyMarkup = markKeyboard
			case ButtonSchedule:
				msg.ReplyMarkup = scheduleKeyboard
			case ButtonFinish:
				msg.Text = "✌️ До встречи"
			case ButtonMarkDay:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: ChildName,
				})

				msg.Text = t.MarksMessageText("Оценки за день: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkWeek:
				ret := t.edu.getEduByWeek(&EduFilter{
					ChildName: ChildName,
				})

				msg.Text = t.MarksMessageText("Оценки за неделю: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkQuarter:
				var keyboards [][]tgbotapi.InlineKeyboardButton
				for _, v := range ListQuarter {
					keyboards = append(keyboards, InlineKeyboardButtonRow(v+" четверть", "select_mark_quarter:"+v))
				}

				keyboards = append(keyboards, BackKeyboardButtonRow(ButtonMarks))

				msg.Text = "Выберите четверть:"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectMarkQuarter:
				quarter := arguments[0]
				ret := t.edu.getEduByQuarter(&EduFilter{
					ChildName: ChildName,
					DiaryType: quarter,
				})

				msg.Text = t.MarksMessageText("Оценки за "+quarter+" четверть: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkSubject:
				subjects := t.edu.getSubjects(&EduFilter{
					ChildName: ChildName,
					DiaryType: Quarter1,
				})

				var keyboards [][]tgbotapi.InlineKeyboardButton
				for i, name := range subjects {
					keyboards = append(keyboards, InlineKeyboardButtonRow(name, "select_mark_subject:"+strconv.Itoa(i)))
				}

				keyboards = append(keyboards, BackKeyboardButtonRow(ButtonMarks))

				msg.Text = "Укажите предмет:"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectMarkSubject:
				i, err := strconv.Atoi(arguments[0])
				if err != nil {
					msg.Text = "Неизвестный предмет"
					break
				}

				subjects := t.edu.getSubjects(&EduFilter{
					ChildName: ChildName,
					DiaryType: Quarter1,
				})

				subject := subjects[i]
				if subject == "" {
					msg.Text = "Неизвестный предмет"
					break
				}

				ret := t.edu.getEduBySubject(&EduFilter{
					ChildName: ChildName,
					Subject:   subject,
				})

				msg.Text = t.MarksMessageText("", []*SchoolSubject{ret})
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarkSubject)
			case ButtonScheduleTomorrow:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: ChildName,
					Date:      time.Now().Add(24 * time.Hour).Format(DDMMYYYY),
				})

				msg.Text = t.ScheduleMessageText("Расписание на завтра: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonScheduleToday:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: ChildName,
					Date:      time.Now().Format(DDMMYYYY),
				})

				msg.Text = t.ScheduleMessageText("Расписание на сегодня: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonScheduleWeek:
				ret := t.edu.getEduByWeek(&EduFilter{
					ChildName: ChildName,
				})

				msg.Text = t.ScheduleWeekMessageText("Расписание за неделю: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonScheduleDate:
				msg.Text = "За какой день, хотите получить расписание (укажите дату в формате d.m.Y):"
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonSelectScheduleDate:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: ChildName,
					Date:      time.Now().Format(DDMMYYYY),
				})

				msg.Text = t.ScheduleMessageText("Расписание на дату: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			}

			msg.ParseMode = "markdown"
			if _, err := t.bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (t *Telegram) getData(data string) (string, map[int]string) {
	splitSata := strings.Split(data, ":")

	first := splitSata[0]
	value := splitSata[1:]

	arguments := make(map[int]string, len(value))
	for k, v := range value {
		arguments[k] = v
	}

	return first, arguments
}

func (t *Telegram) RemoveCallbackMessage(query *tgbotapi.CallbackQuery) {
	t.DeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

func (t *Telegram) DeleteMessage(chatId int64, messageId int) {
	del := tgbotapi.NewDeleteMessage(chatId, messageId)
	t.bot.Send(del)
}

func (t *Telegram) ScheduleWeekMessageText(header string, subjects []*SchoolSubject) string {
	var msgText []string
	var headDay string
	var n int
	for _, v := range subjects {
		if headDay != v.Day {
			headDay = v.Day
			msgText = append(msgText, "\n📚*"+headDay+"*")
			n = 1
		}

		msgText = append(msgText, strconv.Itoa(n)+". "+v.Subject)
		n++
	}

	if msgText != nil {
		return header + "\n" + strings.Join(msgText, "\n")
	} else {
		return "Пусто :("
	}
}

func (t *Telegram) ScheduleMessageText(header string, subjects []*SchoolSubject) string {
	var msgText []string
	for _, v := range subjects {
		var task string
		if task = v.Task; task == "" {
			task = "Нет задания"
		}

		msgText = append(msgText, "✍️\\["+v.Time+"] *"+v.Subject+"*\n📘"+task+"\n")
	}

	if msgText != nil {
		return header + "\n\n" + strings.Join(msgText, "\n")
	} else {
		return "Пусто :("
	}
}

func (t *Telegram) MarksMessageText(header string, subjects []*SchoolSubject) string {
	var msgText []string
	for _, v := range subjects {
		if v.Marks == nil {
			continue
		}

		var marks []string
		for _, m := range v.Marks {
			if m.Reason != "" {
				marks = append(marks, m.Value+"("+m.Reason+")")
			} else {
				marks = append(marks, m.Value)
			}
		}
		msgText = append(msgText, v.Subject+": "+strings.Join(marks, ", "))
	}

	if msgText != nil {
		return header + "\n\n" + strings.Join(msgText, "\n")
	} else {
		return "Оценок нет :("
	}
}

func InlineKeyboardButtonRow(name string, data string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(name, data),
	)
}

func BackKeyboardButtonRow(data string) []tgbotapi.InlineKeyboardButton {
	return InlineKeyboardButtonRow("Назад", data)
}

func BackKeyboardButtonInline(data string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		BackKeyboardButtonRow(data),
	)
}
