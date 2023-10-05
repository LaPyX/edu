package main

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	Quarter1 = "1"
	Quarter2 = "2"
	Quarter3 = "3"
	Quarter4 = "4"
)

const (
	ButtonStart    = "start"
	ButtonRegister = "register"
	ButtonFinish   = "finish"
	ButtonExit     = "exit"
	ButtonCancel   = "cancel"
	ButtonMenu     = "menu"

	ButtonMarks    = "marks"
	ButtonSchedule = "schedule"

	ButtonMarkDay           = "mark_day"
	ButtonMarkWeek          = "mark_week"
	ButtonMarkQuarter       = "mark_quarter"
	ButtonMarkSubject       = "mark_subject"
	ButtonSelectMarkSubject = "select_mark_subject"
	ButtonSelectMarkQuarter = "select_mark_quarter"
	ButtonDoneSubject       = "select_done_subject"

	ButtonScheduleToday      = "schedule_today"
	ButtonScheduleTomorrow   = "schedule_tomorrow"
	ButtonScheduleWeek       = "schedule_week"
	ButtonScheduleDate       = "schedule_date"
	ButtonSelectScheduleDate = "select_schedule_date"
	ButtonChildrenList       = "children_list"
	ButtonSelectChild        = "select_child"
	ButtonSetting            = "settings"
	ButtonUserInfo           = "user_info"
	ButtonUsers              = "users"
	ButtonUpdateUser         = "update_user"
	ButtonSelectUser         = "select_user"
	ButtonDeleteUser         = "delete_user"
	ButtonSelectParentChild  = "select_parent_child"
	ButtonSelectIsParent     = ButtonSelectParentChild + ":parent"
	ButtonSelectIsChild      = ButtonSelectParentChild + ":child"

	ButtonNotify           = "notify_user"
	ButtonNotifyMarksDay   = "notify_marks_day"
	ButtonNotifyMarksDay14 = ButtonNotifyMarksDay + ":" + NotifyMarksDay14
	ButtonNotifyMarksDay18 = ButtonNotifyMarksDay + ":" + NotifyMarksDay18
	ButtonNotifyHomework   = "notify_homework"
	ButtonNotifyHomework16 = ButtonNotifyHomework + ":" + NotifyHomework16
	ButtonNotifyHomework21 = ButtonNotifyHomework + ":" + NotifyHomework21

	ButtonNotifySend         = "notify_user_send"
	ButtonNotifySendMarksDay = ButtonNotifySend + ":" + ButtonNotifyMarksDay
)

const (
	ScheduleHour14 = "14"
	ScheduleHour16 = "16"
	ScheduleHour18 = "18"
	ScheduleHour21 = "21"
)

const (
	RegisterLogin    = "register_login"
	RegisterPassword = "register_password"
)

const (
	NotifyMarksDay14 = "notify_marks_day_" + ScheduleHour14
	NotifyMarksDay18 = "notify_marks_day_" + ScheduleHour18
	NotifyHomework16 = "notify_homework_" + ScheduleHour16
	NotifyHomework21 = "notify_homework_" + ScheduleHour21
)

var ListQuarter = []string{Quarter1, Quarter2, Quarter3, Quarter4}

var UserInfoKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Обновить", ButtonUpdateUser),
	BackKeyboardButtonRow(ButtonSetting),
)

var settingKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Ваши данные", ButtonUserInfo),
	InlineKeyboardButtonRow("Уведомления", ButtonNotify),
	BackKeyboardButtonRow(ButtonStart),
)

var settingAdminKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Ваши данные", ButtonUserInfo),
	InlineKeyboardButtonRow("Пользователи", ButtonUsers),
	InlineKeyboardButtonRow("Уведомления", ButtonNotify),
	BackKeyboardButtonRow(ButtonStart),
)

var notifyKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки за день", ButtonNotifyMarksDay),
	InlineKeyboardButtonRow("Домашнее задание на завтра", ButtonNotifyHomework),
	BackKeyboardButtonRow(ButtonSetting),
)

var notifySendKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки за день", ButtonNotifySendMarksDay),
	BackKeyboardButtonRow(ButtonUsers),
)

var childKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Выбрать учащегося", ButtonChildrenList),
	InlineKeyboardButtonRow("Выйти", ButtonExit),
)

var parentChildKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Я родитель", ButtonSelectIsParent),
	InlineKeyboardButtonRow("Я ребенок", ButtonSelectIsChild),
	InlineKeyboardButtonRow("Выйти", ButtonExit),
)

var registerKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Регистрация", ButtonRegister),
	InlineKeyboardButtonRow("Выйти", ButtonExit),
)

var commandsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки", ButtonMarks),
	InlineKeyboardButtonRow("Расписание", ButtonSchedule),
	InlineKeyboardButtonRow("Учащийся", ButtonChildrenList),
	InlineKeyboardButtonRow("Настройки", ButtonSetting),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
	InlineKeyboardButtonRow("Выйти", ButtonExit),
)

var markKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Оценки за день", ButtonMarkDay),
	InlineKeyboardButtonRow("Оценки за неделю", ButtonMarkWeek),
	InlineKeyboardButtonRow("Оценки за четверть", ButtonMarkQuarter),
	InlineKeyboardButtonRow("Оценки по предмету", ButtonMarkSubject),
	//InlineKeyboardButtonRow("Оценки за прошлую неделю", ButtonMarkSubject),
	BackKeyboardButtonRow(ButtonStart),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
)

var scheduleKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	InlineKeyboardButtonRow("Расписание на сегодня", ButtonScheduleToday),
	InlineKeyboardButtonRow("Расписание на завтра", ButtonScheduleTomorrow),
	InlineKeyboardButtonRow("Расписание за неделю", ButtonScheduleWeek),
	InlineKeyboardButtonRow("Расписание за дату", ButtonScheduleDate),
	//InlineKeyboardButtonRow("Расписание за прошлую неделю", ButtonScheduleDate),
	BackKeyboardButtonRow(ButtonStart),
	InlineKeyboardButtonRow("Завершить", ButtonFinish),
)

type Telegram struct {
	bot  *tgbotapi.BotAPI
	edu  *Edu
	cron *cron.Cron
}

func newTelegram(edu *Edu) *Telegram {
	return &Telegram{
		edu: edu,
	}
}

func (t *Telegram) Run() {
	var err error

	t.cron = cron.New()
	t.bot, err = tgbotapi.NewBotAPIWithClient(os.Getenv("TELEGRAM_APITOKEN"), tgbotapi.APIEndpoint, t.edu.client)
	if err != nil {
		log.Panic(err)
	}

	t.bot.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	log.Printf("Authorized on account %s", t.bot.Self.UserName)

	t.ScheduleStart()
	t.GetUpdatesChan()
}

func (t *Telegram) ScheduleStart() {
	// уведомления об оценках
	// At 14:00 on every day-of-week from Monday through Saturday.
	t.cron.AddFunc("0 14 * * 1-6", t.Notify(NotifyMarksDay14))
	// At 18:00 on every day-of-week from Monday through Saturday.
	t.cron.AddFunc("0 18 * * 1-6", t.Notify(NotifyMarksDay18))

	// уведомления о не выполненной домашке
	// At 16:00 on every day-of-week from Monday through Friday and Sunday.
	t.cron.AddFunc("0 16 * * 1-6", t.Notify(NotifyHomework16))
	// At 21:00 on every day-of-week from Monday through Friday and Sunday.
	t.cron.AddFunc("0 21 * * 1-6", t.Notify(NotifyHomework21))

	t.cron.Start()
}

func (t *Telegram) GetUpdatesChan() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	guard := &Guard{redis: t.edu.redis}
	for update := range updates {

		if update.Message != nil {
			user := guard.auth(update.Message.From)
			t.edu.setCookie(user.Cookie)

			// Construct a new message from the given chat ID and containing
			// the text that we received.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие:")

			if update.Message.IsCommand() {
				// Extract the command from the Message.
				switch update.Message.Command() {
				case ButtonStart:
					if user.IsAuth {
						if len(user.Children) == 0 {
							msg.ReplyMarkup = parentChildKeyboard
						} else {
							msg.Text = user.ChildName + "\n\n" + msg.Text
							msg.ReplyMarkup = commandsKeyboard
						}
					} else {
						msg.ReplyMarkup = registerKeyboard
					}
				case ButtonMenu:
					if user.IsAuth && user.ChildName != "" {
						msg.Text = user.ChildName + "\n\n" + msg.Text
						msg.ReplyMarkup = commandsKeyboard
					} else {
						msg.Text = "Попробуйте начать с этого /start"
					}
				}
			} else if user.Command != "" {
				switch user.Command {
				case RegisterLogin:
					user.Command = RegisterPassword
					//  validate login
					j, _ := json.Marshal(&RegisterEdu{Login: update.Message.Text})
					user.CommandJson = j
					guard.saveUser(user)
					msg.Text = "Введите пароль:"
					msg.ReplyMarkup = CancelKeyboardButtonInline(ButtonRegister)
				case RegisterPassword:
					var regedu *RegisterEdu
					err := json.Unmarshal(user.CommandJson, &regedu)
					if err == nil {
						// validate password
						regedu.Password = update.Message.Text
						cookie, err := t.edu.loginRequest(regedu)
						if err == nil {
							msg.Text = "Регистрация прошла успешно.\n\nВыберите действие:"
							user.Command = ""
							user.CommandJson = nil
							user.IsAuth = true
							user.Cookie = cookie
							user.Children = nil
							user.ChildName = ""
							user.LoginEdu = regedu.Login
							guard.saveUser(user)
							msg.ReplyMarkup = parentChildKeyboard
						} else {
							msg.Text = err.Error() + "\n\nВведите пароль:"
							msg.ReplyMarkup = CancelKeyboardButtonInline(ButtonRegister)
						}
					} else {
						fmt.Println(err)
						msg.Text = "Ошибка, извините мы уже работаем над ее решением"
						msg.ReplyMarkup = CancelKeyboardButtonInline(ButtonRegister)
					}
				}
			}

			// Send the message.
			if _, err := t.bot.Send(msg); err != nil {
				panic(err)
			}
		} else if update.CallbackQuery != nil {
			user := guard.auth(update.CallbackQuery.From)
			t.edu.setCookie(user.Cookie)

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, user.ChildName+"\n\n")
			t.RemoveCallbackMessage(update.CallbackQuery)
			data, arguments := t.getData(update.CallbackQuery.Data)
			if t.bot.Debug {
				fmt.Println("arguments", arguments)
			}

			if !user.IsAuth && data != ButtonRegister {
				data = ButtonCancel
			}

			switch data {
			case ButtonNotifySend:
				notifySend := arguments[0]
				if notifySend == ButtonNotifyMarksDay {
					t.SendMarksDayNotification(user, NotifyMarksDay14)
				}
				msg.Text = "Уведомления: \n"
				msg.ReplyMarkup = notifySendKeyboard
			case ButtonNotify:
				msg.Text = "Уведомления: \n"
				msg.ReplyMarkup = notifyKeyboard
			case ButtonNotifyMarksDay:
				notifyName := arguments[0]
				if notifyName != "" {
					user.SyncNotification(notifyName)
					guard.saveUser(user)
				}
				msg.Text = "Выберите в какое время хотите получать уведомление: \n"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					InlineKeyboardButtonRow("14:00 ("+boolToText(user.HasNotification(NotifyMarksDay14), "Включено", "Отключено")+")", ButtonNotifyMarksDay14),
					InlineKeyboardButtonRow("18:00 ("+boolToText(user.HasNotification(NotifyMarksDay18), "Включено", "Отключено")+")", ButtonNotifyMarksDay18),
					BackKeyboardButtonRow(ButtonNotify),
				)
			case ButtonNotifyHomework:
				notifyName := arguments[0]
				if notifyName != "" {
					user.SyncNotification(notifyName)
					guard.saveUser(user)
				}
				msg.Text = "Выберите в какое время хотите получать уведомление: \n"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					InlineKeyboardButtonRow("16:00 ("+boolToText(user.HasNotification(NotifyHomework16), "Включено", "Отключено")+")", ButtonNotifyHomework16),
					InlineKeyboardButtonRow("21:00 ("+boolToText(user.HasNotification(NotifyHomework21), "Включено", "Отключено")+")", ButtonNotifyHomework21),
					BackKeyboardButtonRow(ButtonNotify),
				)
			case ButtonSetting:
				msg.Text = "Настройки: \n"
				if user.isAdmin() {
					if dt := lastModifiedApp(); dt != nil {
						msg.Text = fmt.Sprintf(
							"Последнее обновление: %s \n\n %s",
							dt.Format("2006-01-02 15:04:05"),
							msg.Text,
						)
					}
					msg.ReplyMarkup = settingAdminKeyboard
				} else {
					msg.ReplyMarkup = settingKeyboard
				}
			case ButtonUserInfo:
				msg.Text = t.UserInfo(user)
				msg.ReplyMarkup = UserInfoKeyboard
			case ButtonUpdateUser:
				user.fill(update.CallbackQuery.From)
				parent, _ := t.findParentUser(user)
				if parent != nil {
					user.ParentId = parent.Id
				}
				guard.saveUser(user)
				msg.Text = t.UserInfo(user)
				msg.ReplyMarkup = UserInfoKeyboard
			case ButtonRegister:
				user.Command = RegisterLogin
				guard.saveUser(user)
				msg.Text = "Введите логин:"
				msg.ReplyMarkup = CancelKeyboardButtonInline(ButtonCancel)
			case ButtonUsers:
				keys, _ := t.edu.redis.Keys(PrefixUser + ":*").Result()
				var keyboards [][]tgbotapi.InlineKeyboardButton
				for _, key := range keys {
					var user *User
					user, err := t.getUser(key)
					if err != nil {
						continue
					}

					name := fmt.Sprintf("%s %s(Id: %s)", user.FirstName, user.LastName, user.Id)
					keyboards = append(keyboards, InlineKeyboardButtonRow(name, "select_user:"+user.Id))
				}

				keyboards = append(keyboards, BackKeyboardButtonRow(ButtonSetting))

				msg.Text = fmt.Sprintf("Пользователи (%s): \n", strconv.Itoa(len(keyboards)-1))
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectUser:
				id := arguments[0]
				user := guard.find(arguments[0])

				if user == nil {
					msg.Text = fmt.Sprintf("Пользователь ID%s не найден", id)
					msg.ReplyMarkup = BackKeyboardButtonInline(ButtonUsers)
					break
				}
				msg.Text = t.UserInfo(user)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					//InlineKeyboardButtonRow("Отправить уведомление", ButtonNotifySend),
					InlineKeyboardButtonRow("Удалить", fmt.Sprintf("%s:%s", ButtonDeleteUser, id)),
					BackKeyboardButtonRow(ButtonUsers),
				)
			case ButtonDeleteUser:
				id := arguments[0]
				guard.removeUser(id)
				msg.Text = fmt.Sprintf("Пользователь ID:%s удален.", id)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonUsers)
			case ButtonExit:
				guard.logout(user)
				msg.Text = "✌️ До встречи"
			case ButtonCancel:
				guard.logout(user)
				msg.Text = "Выберите действие:"
				msg.ReplyMarkup = registerKeyboard
			case ButtonStart:
				msg.Text += "Выберите действие:"
				msg.ReplyMarkup = commandsKeyboard
			case ButtonMarks:
				msg.Text += "Выберите действие:"
				msg.ReplyMarkup = markKeyboard
			case ButtonSchedule:
				msg.Text += "Выберите действие:"
				msg.ReplyMarkup = scheduleKeyboard
			case ButtonSelectParentChild:
				if arguments[0] == "parent" {
					user.IsChildren = false
					user.ParentId = ""
				} else {
					user.IsChildren = true
					parent, _ := t.findParentUser(user)
					if parent != nil {
						user.ParentId = parent.Id
					}
				}

				guard.saveUser(user)
				if len(user.Children) == 0 {
					msg.Text = "Выберите действие:"
					msg.ReplyMarkup = childKeyboard
				} else {
					msg.Text = user.ChildName + "\n\nВыберите действие:"
					msg.ReplyMarkup = commandsKeyboard
				}
			case ButtonSelectIsParent:
				user.IsChildren = true
				guard.saveUser(user)
				if len(user.Children) == 0 {
					msg.ReplyMarkup = childKeyboard
				} else {
					msg.Text = user.ChildName + "\n\nВыберите действие:"
					msg.ReplyMarkup = commandsKeyboard
				}
			case ButtonChildrenList:
				user.Children = t.edu.getChildren(&EduFilter{})
				guard.saveUser(user)

				var keyboards [][]tgbotapi.InlineKeyboardButton
				for i, name := range user.Children {
					keyboards = append(keyboards, InlineKeyboardButtonRow(name, "select_child:"+strconv.Itoa(i)))
				}

				keyboards = append(keyboards, BackKeyboardButtonRow(ButtonMarks))

				msg.Text = "Выберите учащегося:"
				if len(user.Children) == 0 {
					msg.Text += "\n\nСписок учащихся пуст :("
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectChild:
				childKey, err := strconv.Atoi(arguments[0])
				if err != nil {
					msg.Text = "Неизвестный аргумент: " + arguments[0]
					break
				}

				childName := user.Children[childKey]
				user.ChildName = childName
				guard.saveUser(user)
				msg.Text = user.ChildName + "\n\nВыберите действие:"
				msg.ReplyMarkup = commandsKeyboard
			case ButtonFinish:
				msg.Text = "✌️ До встречи"
			case ButtonMarkDay:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: user.ChildName,
				})

				msg.Text += t.MarksMessageText("Оценки за день: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkWeek:
				ret := t.edu.getEduByWeek(&EduFilter{
					ChildName: user.ChildName,
				})

				msg.Text += t.MarksMessageText("Оценки за неделю: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkQuarter:
				keyboards := t.KeyBoardQuarterList(ButtonSelectMarkQuarter, ButtonMarks)
				msg.Text = "Выберите четверть:"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectMarkQuarter:
				quarter := arguments[0]
				ret := t.edu.getEduByQuarter(&EduFilter{
					ChildName: user.ChildName,
					DiaryType: quarter,
				})

				msg.Text += t.MarksMessageText("Оценки за "+quarter+" четверть: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarks)
			case ButtonMarkSubject:
				quarter := arguments[0]
				if quarter == "" {
					keyboards := t.KeyBoardQuarterList(ButtonMarkSubject, ButtonMarks)
					msg.Text = "Выберите четверть:"
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
					break
				}

				subjects := t.edu.getSubjects(&EduFilter{
					ChildName: user.ChildName,
					DiaryType: quarter,
				})

				var keyboards [][]tgbotapi.InlineKeyboardButton
				for i, name := range subjects {
					keyboards = append(keyboards, InlineKeyboardButtonRow(name, ButtonSelectMarkSubject+":"+quarter+":"+strconv.Itoa(i)))
				}

				keyboards = append(keyboards, BackKeyboardButtonRow(ButtonMarkSubject))

				msg.Text = "Укажите предмет:"
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)
			case ButtonSelectMarkSubject:
				quarter := arguments[0]
				i, err := strconv.Atoi(arguments[1])
				if err != nil {
					msg.Text = "Неизвестный предмет"
					break
				}

				subjects := t.edu.getSubjects(&EduFilter{
					ChildName: user.ChildName,
					DiaryType: quarter,
				})

				subject := subjects[i]
				if subject == "" {
					msg.Text = "Неизвестный предмет"
					break
				}

				ret := t.edu.getEduBySubject(&EduFilter{
					ChildName: user.ChildName,
					Subject:   subject,
					DiaryType: quarter,
				})

				msg.Text += t.MarksMessageText("Оценки за "+quarter+" четверть: ", []*SchoolSubject{ret})
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonMarkSubject)
			case ButtonDoneSubject:
				msg, _ = t.getScheduleQuery(user, user.ChildName, msg, arguments[0], toString(arguments[1]))
			case ButtonScheduleTomorrow:
				msg, _ = t.getScheduleQuery(user, user.ChildName, msg, data, nil)
			case ButtonScheduleToday:
				msg, _ = t.getScheduleQuery(user, user.ChildName, msg, data, nil)
			case ButtonScheduleWeek:
				ret := t.edu.getEduByWeek(&EduFilter{
					ChildName: user.ChildName,
				})

				msg.Text += t.ScheduleWeekMessageText("Расписание за неделю: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonScheduleDate:
				msg.Text = "За какой день, хотите получить расписание (укажите дату в формате d.m.Y):"
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			case ButtonSelectScheduleDate:
				ret := t.edu.getEduByDay(&EduFilter{
					ChildName: user.ChildName,
					Date:      time.Now().Format(DDMMYYYY),
				})

				msg.Text += t.ScheduleMessageText("Расписание на дату: ", ret)
				msg.ReplyMarkup = BackKeyboardButtonInline(ButtonSchedule)
			}

			msg.ParseMode = "Markdown"
			if _, err := t.bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (t *Telegram) getScheduleQuery(user *User, ChildName string, msg tgbotapi.MessageConfig, data string, subjectIndex *string) (tgbotapi.MessageConfig, bool) {
	var (
		date  string
		title string
		ret   []*SchoolSubject
	)
	switch data {
	case ButtonScheduleTomorrow:
		date = time.Now().Add(24 * time.Hour).Format(DDMMYYYY)
		title = "Расписание на завтра: "
	case ButtonScheduleToday:
		date = time.Now().Format(DDMMYYYY)
		title = "Расписание на сегодня: "
	default:
		return msg, false
	}

	if subjectIndex != nil {
		t.edu.SaveHomework(ChildName, date, *subjectIndex)
	}

	ret = t.edu.getEduByDay(&EduFilter{
		ChildName: ChildName,
		Date:      date,
	})

	var keyboards [][]tgbotapi.InlineKeyboardButton

	allDone := len(keyboards) == 1
	msg.Text += t.ScheduleMessageText(title, ret)

	if !allDone && user.IsChildren {
		msg.Text += "\n *Укажи по каким предметам выполнил(а) Д/З:* "
		for i, subj := range ret {
			if subj.IsDone || subj.Task == "" {
				continue
			}
			keyboards = append(keyboards, InlineKeyboardButtonRow(subj.Subject+" 👍", ButtonDoneSubject+":"+data+":"+strconv.Itoa(i)))
		}
	}

	keyboards = append(keyboards, BackKeyboardButtonRow(ButtonSchedule))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboards...)

	return msg, allDone
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
		var (
			task      = "-"
			doneSmile = "✅"
		)

		if v.Task != "" {
			task = v.Task
			doneSmile = boolToText(v.IsDone, "✅", "😡")
		}

		msgText = append(msgText, "✍️\\["+v.Time+"] *"+v.Subject+"* "+doneSmile+"\n📘"+t.QuoteMeta(task, "_")+"\n")
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

		text := "✍*" + v.Subject + "*: " + strings.Join(marks, ", ")
		if v.AvgMark != "" {
			text += "\nСредняя: " + v.AvgMark
			if v.Total != "" {
				text += ", Итог: " + v.AvgMark
			}
		}

		msgText = append(msgText, text)
	}

	if msgText != nil {
		if header != "" {
			header += "\n\n"
		}
		return header + strings.Join(msgText, "\n\n")
	} else {
		return "Оценок нет :("
	}
}

func (t *Telegram) UserInfo(user *User) string {
	header := fmt.Sprintf("Данные о пользователе ID: %s", user.Id)

	if user.IsChildren {
		header += " - (Ребёнок)"
	} else {
		header += " - (Родитель)"
	}

	body := fmt.Sprintf("ID: %s \n", user.Id)
	body += fmt.Sprintf("UserName: @%s \n", t.QuoteMeta(user.UserName))
	body += fmt.Sprintf("Имя: %s \n", t.QuoteMeta(user.FirstName))
	body += fmt.Sprintf("Фамилия: %s \n", t.QuoteMeta(user.LastName))
	if !user.IsChildren {
		body += fmt.Sprintf("Дети: %s \n", t.QuoteMeta(strings.Join(user.Children, ", ")))
	} else if user.ParentId != "" {
		parent, _ := t.getUserById(user.ParentId)
		if parent != nil {
			body += fmt.Sprintf("Родитель: @%s (%s %s ID:%s) \n", t.QuoteMeta(parent.UserName), t.QuoteMeta(parent.FirstName), t.QuoteMeta(parent.LastName), parent.Id)
		}
	}
	body += fmt.Sprintf("Выбранный учащийся: %s", t.QuoteMeta(user.ChildName))

	return header + "\n\n" + body
}

func (t *Telegram) findParentUser(u *User) (*User, error) {
	keys, _ := t.edu.redis.Keys(PrefixUser + ":*").Result()
	for _, key := range keys {
		user, _ := t.getUser(key)
		if user != nil && u.Id != user.Id && user.LoginEdu == u.LoginEdu && !user.IsChildren {
			return user, nil
		}
	}
	return nil, errors.New("Parent not found")
}

func (t *Telegram) getUserById(id string) (*User, error) {
	return t.getUser(KeyUser(id))
}

func (t *Telegram) getUser(prefixKey string) (*User, error) {
	var user *User
	u, _ := t.edu.redis.Get(prefixKey).Bytes()
	err := json.Unmarshal(u, &user)
	if err != nil {
		return nil, err
	}
	if user.Id == "" {
		return nil, errors.New("user id: " + prefixKey + " not found")
	}
	return user, nil
}

func (t *Telegram) getUsers() []*User {
	var users []*User
	keys, _ := t.edu.redis.Keys(PrefixUser + ":*").Result()
	for _, key := range keys {
		var user *User
		user, err := t.getUser(key)
		if err != nil {
			continue
		}

		users = append(users, user)
	}
	return users
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

func CancelKeyboardButtonInline(data string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		InlineKeyboardButtonRow("Отменить", data),
	)
}

func (t *Telegram) QuoteMeta(s string, spec ...string) string {
	specs := `_-\.+*?()|[]{}^$`
	if spec != nil {
		specs = strings.Join(spec, "")
	}
	var specialBytes [16]byte
	for _, b := range []byte(specs) {
		specialBytes[b%16] |= 1 << (b / 16)
	}

	special := func(b byte) bool {
		return b < utf8.RuneSelf && specialBytes[b%16]&(1<<(b/16)) != 0
	}
	// A byte loop is correct because all metacharacters are ASCII.
	var i int
	for i = 0; i < len(s); i++ {
		if special(s[i]) {
			break
		}
	}
	// No meta characters found, so return original string.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)
	copy(b, s[:i])
	j := i
	for ; i < len(s); i++ {
		if special(s[i]) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[:j])
}

func (t *Telegram) KeyBoardQuarterList(buttonAction string, buttonBack string) [][]tgbotapi.InlineKeyboardButton {
	var keyboards [][]tgbotapi.InlineKeyboardButton
	for _, v := range ListQuarter {
		keyboards = append(keyboards, InlineKeyboardButtonRow(v+" четверть", buttonAction+":"+v))
	}
	return append(keyboards, BackKeyboardButtonRow(buttonBack))
}

func (t *Telegram) Notify(notifyName string) func() {
	return func() {
		for _, user := range t.getUsers() {
			if user.Notification != nil {
				for _, notify := range user.Notification {
					if notify == notifyName && user.Children != nil {
						switch notifyName {
						case NotifyMarksDay14, NotifyMarksDay18:
							t.SendMarksDayNotification(user, notifyName)
						case NotifyHomework16, NotifyHomework21:
							t.SendHomeworkNotification(user, notifyName)
						}

						break
					}
				}
			}
		}
	}
}

func (t *Telegram) SendMarksDayNotification(user *User, notifyName string) {
	id, _ := strconv.ParseInt(user.Id, 10, 64)
	msg := tgbotapi.NewMessage(id, "")
	msg.ParseMode = "markdown"

	t.edu.setCookie(user.Cookie)

	// TODO create template by notifyName
	for _, child := range user.Children {
		if user.IsChildren && child != user.ChildName {
			continue
		}

		ret := t.edu.getEduByDay(&EduFilter{
			ChildName: child,
		})

		msg.Text += "Оценки за день (" + child + "): \n\n"
		msg.Text += strings.Trim(t.MarksMessageText("", ret), "\n\n")
		msg.Text += "\n\n"
	}

	msg.Text = strings.Trim(msg.Text, "\n\n")
	t.bot.Send(msg)
	log.Println("Schedule notify: " + notifyName + " send from " + user.Id)
}

func (t *Telegram) SendHomeworkNotification(user *User, notifyName string) {
	id, _ := strconv.ParseInt(user.Id, 10, 64)
	msg := tgbotapi.NewMessage(id, "")
	msg.ParseMode = "markdown"

	t.edu.setCookie(user.Cookie)

	// TODO create template by notifyName
	for _, child := range user.Children {
		if user.IsChildren && child != user.ChildName {
			continue
		}

		msgNew, isAllDone := t.getScheduleQuery(user, child, msg, ButtonScheduleTomorrow, nil)
		if isAllDone {
			if !user.IsChildren {
				msg = msgNew
				msg.Text = "✅ Домашнее задание выполнено (" + child + "): \n\n" + msg.Text
			}

			continue
		}

		msg = msgNew
		msg.Text = "😡 Домашнее задание на завтра (" + child + ") не выполнено: \n\n" + msg.Text
	}

	if msg.Text == "" {
		log.Println("Schedule notify: все ДЗ выполнены уведомление не отправляем")
		return
	}

	t.bot.Send(msg)
	log.Println("Schedule notify: " + notifyName + " send from " + user.Id)
}

func boolToText(b bool, isTrue string, isFalse string) string {
	if b {
		return isTrue
	}
	return isFalse
}

func toString(str string) *string {
	return &str
}
