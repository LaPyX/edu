package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Mark struct {
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type SchoolSubject struct {
	Day     string  `json:"day"`
	Time    string  `json:"time"`
	Subject string  `json:"subject"`
	Task    string  `json:"task"`
	Comment string  `json:"comment"`
	Marks   []*Mark `json:"marks"`
	Total   string  `json:"total"`
	AvgMark string  `json:"avg_mark"`
}

type EduFilter struct {
	ChildName string `json:"child_name"`
	DiaryType string `json:"diary_type"`
	Date      string `json:"date"`
	Subject   string `json:"subject"`
}

type Edu struct {
	client          *http.Client
	cookie          []*http.Cookie
	quarterSubjects map[string][]*SchoolSubject
	redis           *redis.Client
}

func newEdu() *Edu {
	edu := &Edu{
		client: newClient(),
		cookie: checkAuth(),
		//redis:  newRedis(),
	}

	if edu.cookie == nil {
		edu.cookie = edu.loginRequest("9046762614", os.Getenv("EDU_PASSWORD"))
	}

	return edu
}

func checkAuth() []*http.Cookie {
	// TODO redis
	b, err := os.ReadFile("cookie.cache")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var cookie []*http.Cookie
	err = json.Unmarshal(b, &cookie)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return cookie
}

func newClient() *http.Client {
	jar := NewJar()

	return &http.Client{
		Jar: jar,
	}
}

func newRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return client
}

func queryDoc(data string) *goquery.Document {
	node, err := html.Parse(strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	doc := goquery.NewDocumentFromNode(node)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func saveToFile(path string, text string) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(text)
}

func (edu *Edu) eduRequest(filter *EduFilter) []byte {
	urlUslugi := "https://uslugi.tatarstan.ru/edu"
	method := "POST"

	payload := &bytes.Buffer{}

	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("child_name", filter.ChildName)
	_ = writer.WriteField("diary_type", filter.DiaryType)
	_ = writer.WriteField("date", filter.Date)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req, err := http.NewRequest(method, urlUslugi, payload)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	edu.client.Jar.SetCookies(req.URL, edu.cookie)
	res, err := edu.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println(res.StatusCode)

	return body
}

func (edu *Edu) loginRequest(login string, password string) []*http.Cookie {
	fmt.Println("Login auth")
	urlUslugi := "https://uslugi.tatarstan.ru/user/login"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("user_login_form_model[phone_number]", login)
	_ = writer.WriteField("user_login_form_model[password]", password)
	_ = writer.WriteField("user_login_form_model[remember_me]", "1")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req, err := http.NewRequest(method, urlUslugi, payload)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := edu.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	cookie := edu.client.Jar.Cookies(req.URL)
	j, _ := json.Marshal(cookie)

	// TODO redis
	saveToFile("cookie.cache", string(j))

	return cookie
}

func (edu *Edu) getEduByWeek(filter *EduFilter) []*SchoolSubject {
	fmt.Println("Get estimates by weeks")

	filter.DiaryType = "diary"
	doc := edu.eduRequest(filter)

	if doc != nil {
		var schoolSubjects []*SchoolSubject
		doc := queryDoc(string(doc))

		var day string
		rMark, _ := regexp.Compile("([0-9]|б|н)")
		rDays, _ := regexp.Compile("(Понедельник|Вторник|Среда|Четверг|Пятница|Суббота)|([0-9]+)")
		doc.Find("div.tt tbody tr").Each(func(_ int, tr *goquery.Selection) {
			if d := tr.Find(".tt-days").First(); d.Text() != "" {
				day = strings.Join(rDays.FindAllString(d.Text(), -1), " ")
			}

			subj := tr.Find(".tt-subj").First()
			subject := strings.TrimSpace(subj.Text())
			if subject == "" {
				return
			}

			task := tr.Find(".tt-task").First()
			mark := tr.Find(".tt-mark").First()

			var marks []*Mark
			for _, v := range rMark.FindAllString(mark.Text(), -1) {
				marks = append(marks, &Mark{Value: v})
			}

			sd := &SchoolSubject{
				Day:     day,
				Subject: subject,
				Task:    strings.TrimSpace(task.Text()),
				Marks:   marks,
			}

			schoolSubjects = append(schoolSubjects, sd)
		})

		return schoolSubjects
	}

	return nil
}

func (edu *Edu) getEduByDay(filter *EduFilter) []*SchoolSubject {
	fmt.Println("Get estimates by day: " + filter.Date)

	filter.DiaryType = "day"

	doc := edu.eduRequest(filter)
	if doc != nil {
		var schoolSubjects []*SchoolSubject
		doc := queryDoc(string(doc))

		doc.Find("table.extra-table tbody tr").Each(func(_ int, tr *goquery.Selection) {
			timeText := tr.Find("td").Eq(0).Text()
			subject := tr.Find("td").Eq(1).Text()
			task := tr.Find("td").Eq(2).Text()
			comment := tr.Find("td").Eq(3).Text()

			if subject == "" {
				return
			}

			var marks []*Mark
			tr.Find("td").Eq(4).Find(".tooltip-sts").Each(func(_ int, div *goquery.Selection) {
				reason := div.Find(".tooltip-sts-content").First().Text()
				mark := div.Next().Text()

				marks = append(marks, &Mark{
					Value:  mark,
					Reason: strings.TrimSpace(reason),
				})
			})

			sd := &SchoolSubject{
				Day:     filter.Date,
				Time:    strings.TrimSpace(timeText),
				Subject: strings.TrimSpace(subject),
				Task:    strings.TrimSpace(task),
				Comment: strings.TrimSpace(comment),
				Marks:   marks,
			}

			schoolSubjects = append(schoolSubjects, sd)
		})

		return schoolSubjects
	}

	return nil
}

func (edu *Edu) getEduByQuarter(filter *EduFilter) []*SchoolSubject {
	// TODO redis
	if edu.quarterSubjects[filter.DiaryType] != nil {
		return edu.quarterSubjects[filter.DiaryType]
	}

	fmt.Println("Get estimates by quarter: " + filter.DiaryType)
	doc := edu.eduRequest(filter)
	if doc != nil {
		var schoolSubjects []*SchoolSubject
		doc := queryDoc(string(doc))

		rMark, _ := regexp.Compile("([0-9])")
		trAll := doc.Find("table.extra-table tbody tr")
		trAll.Each(func(i int, tr *goquery.Selection) {
			if (i + 1) == trAll.Length() {
				return
			}

			subject := tr.Find("td").Eq(0).Text()
			avg := tr.Find("td").Eq(2).Text()
			total := tr.Find("td").Eq(4).Text()

			marksAll := rMark.FindAllString(tr.Find("td").Eq(1).Text(), -1)
			marks := strings.Join(marksAll, " ")

			sd := &SchoolSubject{
				Day:     filter.DiaryType + " четверть",
				Subject: strings.TrimSpace(subject),
				Marks:   []*Mark{{Value: marks}},
				Total:   total,
				AvgMark: avg,
			}

			schoolSubjects = append(schoolSubjects, sd)
		})

		edu.quarterSubjects[filter.DiaryType] = schoolSubjects

		return schoolSubjects
	}

	return nil
}

func (edu *Edu) getEduBySubject(filter *EduFilter) *SchoolSubject {
	fmt.Println("Get estimates by quarter: " + filter.DiaryType + ", subject: " + filter.Subject)

	subjects := edu.getEduByQuarter(filter)

	for _, v := range subjects {
		if filter.Subject == v.Subject {
			return v
		}
	}

	return nil
}

func (edu *Edu) getSubjects(filter *EduFilter) map[int]string {
	fmt.Println("Get subjects")

	var subjects = make(map[int]string)
	for k, v := range edu.getEduByQuarter(filter) {
		subjects[k] = v.Subject
	}

	return subjects
}
