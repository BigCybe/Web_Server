package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

var tokenEncodeString string = "something"

//Получение токена
func GetToken(name, password, id string) (string, error) {
	token := jwt.New(&jwt.SigningMethodHS256{})
	token.Claims["username"] = name
	token.Claims["password"] = password
	token.Claims["userid"] = id
	return token.SignedString([]byte(tokenEncodeString))
}



// Проверка правильности полученных данных
func IsExistInBd(Login, Password string) (bool, string) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()

	rows, err := db.Query("select login, password, user_id from servusers")
	CheckError(err)
	defer rows.Close()

	flagISExist := false
	var id int
	idUser := 0
	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Login, &u.Password, &id)
		CheckError(err)
		if u.Login == Login && u.Password == Hash(Password) {
			flagISExist = true
			idUser = id
		}
	}
	return flagISExist, string(idUser)
}

func IsExistInBdName(Name string) bool {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()

	rows, err := db.Query("select name from servusers")
	CheckError(err)
	defer rows.Close()

	flagISExist := false
	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Name)
		CheckError(err)
		if u.Name == Name {
			flagISExist = true
		}
	}
	return flagISExist
}
