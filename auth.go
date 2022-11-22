package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"strconv"
)

const PrefixUser = "users"

type Guard struct {
	redis *redis.Client
}

type User struct {
	Id           string         `json:"id"`
	LoginEdu     string         `json:"login_edu"`
	FirstName    string         `json:"first_name,omitempty"`
	LastName     string         `json:"last_name,omitempty"`
	UserName     string         `json:"username"`
	IsAuth       bool           `json:"is_auth"`
	Cookie       []*http.Cookie `json:"cookie"`
	IsChildren   bool           `json:"is_children"`
	Children     []string       `json:"children"`
	Command      string         `json:"command"`
	CommandJson  []byte         `json:"command_json"`
	ChildName    string         `json:"child_name"`
	ParentId     string         `json:"parent_id"`
	Notification []string       `json:"notification"`
}

func (u *User) HasNotification(notifyName string) bool {
	for _, name := range u.Notification {
		if name == notifyName {
			return true
		}
	}
	return false
}

func (u *User) SyncNotification(notifyName string) {
	var notify []string
	if !u.HasNotification(notifyName) {
		notify = append(notify, notifyName)
	}
	for _, name := range u.Notification {
		if name == notifyName {
			continue
		}
		notify = append(notify, name)
	}
	u.Notification = notify
}

func (u *User) isAdmin() bool {
	return u.Id == AdminId
}

func (g *Guard) auth(tgUser *tgbotapi.User) *User {
	user := g.findOrCreateUser(tgUser)

	//fmt.Println(user)

	return user
}

func (g *Guard) find(id string) *User {
	var user *User
	j, err := g.redis.Get(KeyUser(id)).Bytes()
	err = json.Unmarshal(j, &user)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return user
}

func (g *Guard) findOrCreateUser(tgUser *tgbotapi.User) *User {
	var user *User
	id := strconv.FormatInt(tgUser.ID, 10)
	user = g.find(id)
	if user == nil {
		user := &User{
			Id:         id,
			FirstName:  tgUser.FirstName,
			LastName:   tgUser.LastName,
			UserName:   tgUser.UserName,
			IsAuth:     false,
			Cookie:     nil,
			IsChildren: false,
			Children:   nil,
		}
		g.saveUser(user)
		return user
	}
	return user
}

func (g *Guard) saveUser(u *User) bool {
	j, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = g.redis.Set(KeyUser(u.Id), j, 0).Err()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func KeyUser(id string) string {
	return fmt.Sprintf("%s:%s", PrefixUser, id)
}

func (g *Guard) logout(u *User) {
	u.IsAuth = false
	u.Command = ""
	u.CommandJson = nil
	u.Cookie = nil
	g.saveUser(u)
}

func (g *Guard) removeUser(id string) {
	g.redis.Del(KeyUser(id))
}

func (u *User) fill(tgUser *tgbotapi.User) *User {
	u.FirstName = tgUser.FirstName
	u.LastName = tgUser.LastName
	u.UserName = tgUser.UserName
	return u
}
