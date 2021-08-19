package main

import "time"

// Данные пользователя
type User struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Данные группы
type Group struct {
	GroupName        string `json:"group_name"`
	GroupDescription string `json:"group_description"`
	ParentId         int `json:"parent_id"`
	GroupId          int `json:"group_id"`
}

type Task struct {
	TaskId      string `json:"task_id"`
	GroupId     int `json:"group_id"`
	Task        string `json:"task"`
	Completed   bool `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// Данные для входа в БД
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "874874874"
	dbname   = "server"
)

type Writer interface {
	Write(p []byte) (n int, err error)
}
