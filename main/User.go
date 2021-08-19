package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
)

//Хэширование пароля
func Hash(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}

// Валидация данных пользователя
func (u User) ValidateUser() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.Required, validation.Length(1, 50)),
		validation.Field(&u.Login, validation.Required, validation.Length(1, 50)),
		validation.Field(&u.Password, validation.Required, validation.Length(7, 20), validation.Match(regexp.MustCompile("[A-Z]")).Error("need at least one appear case"), validation.Match(regexp.MustCompile("[1-9]")).Error("must be digits")))
}

func (u *User) LoggingUser(){
	log.WithFields(log.Fields{
		"name": u.Name,
		"login": u.Login,
		"password": u.Password,
	}).Info("")
}

// /auth/new Обаробка запроса создания нового пользователя
func HandlePostNewUserEcho(c echo.Context) error{
	log.Println("Begin create new user")
	var u User
	t, _ := ioutil.ReadAll(c.Request().Body)
	json.Unmarshal(t, &u)
	body, err := ioutil.ReadAll(c.Request().Body)
	CheckError(err)
	json.Unmarshal(body, &u)
	err = u.ValidateUser()
	Exist := IsExistInBdName(u.Name)
	if err == nil {
		if !Exist {
			UserInBd(u.Name, u.Login, Hash(u.Password))
			log.Println("New user successfully completed")
			u.LoggingUser()
			u.Password = ""
			b, _ := json.Marshal(u)
			return c.String(http.StatusCreated, string(b))
		} else {
			log.Println("This name already taken")
			u.LoggingUser()
			return c.String(http.StatusAccepted, "This name already taken")
		}
	} else {
		CheckError(err)
		log.Println("Create new user failed")
		u.LoggingUser()
		return c.String(http.StatusBadRequest, err.Error())
	}
}

// /login Обработка запроса получения токена
func HandlePostTokenEcho(c echo.Context) error{
	log.Println("Calculate token")
	var u User
	body, _ := ioutil.ReadAll(c.Request().Body)
	json.Unmarshal(body, &u)
	fmt.Println(string(body))
	isExist, id := IsExistInBd(u.Login, u.Password)
	if isExist {
		b, _ := GetToken(u.Name, u.Password, id)
		log.Println("Token successfully getting")
		return c.String(http.StatusOK, b)
	} else {
		log.Println("Calculated token crashed. User not found")
		return c.String(http.StatusAccepted, "User not found")
	}
}

//Запись нового пользователя в базу данных (Таблица 'servusers')
func UserInBd(Name, Login, Password string) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	defer db.Close()
	insertStmt := `insert into  "servusers"("name", "login", "password") values($1, $2, $3)` //Запись данных в БД
	_, e := db.Exec(insertStmt, Name, Login, Password)
	CheckError(e)
	err = db.Ping()
	CheckError(err)
}


