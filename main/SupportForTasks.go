package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TaskByName []Task

func (a TaskByName) Len() int           { return len(a) }
func (a TaskByName) Less(i, j int) bool { return a[i].Task < a[j].Task }
func (a TaskByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func SortTaskByName(Tasks[] Task)  []Task {
	sort.Sort(TaskByName(Tasks))
	return Tasks
}

type TaskByGroup []Task

func (a TaskByGroup) Len() int           { return len(a) }
func (a TaskByGroup) Less(i, j int) bool { return a[i].GroupId < a[j].GroupId }
func (a TaskByGroup) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func SortTaskByGroup(Tasks[] Task)  []Task {
	sort.Sort(TaskByGroup(Tasks))
	return Tasks
}

func HashT (str string) string {
	s := sha1.New()
	s.Write([]byte(str))
	sha := hex.EncodeToString(s.Sum(nil))
	return sha
}

// Запись задачи в БД
func TaskToBD(t Task){
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	defer db.Close()
	createdAt := FormatDate(t.CreatedAt)
	completedAt := FormatDate(t.CompletedAt)
	insertStmt := `insert into  "servtasks"("task_id", "group_id", "task", "completed", "created_at", "completed_at") values($1, $2, $3, $4, $5, $6)` //Запись данных в БД
	_, e := db.Exec(insertStmt, t.TaskId, t.GroupId, t.Task, t.Completed, createdAt, completedAt)
	CheckError(e)
	err = db.Ping()
	CheckError(err)
}

// Форматирование даты
func FormatDate (Time time.Time) string {
	return Time.Format("2006-01-02 15:04:05.0+00")
}

// Получение списка групп из БД
func GetTasksFromBD() []Task{
	var t Task
	var Tasks[] Task
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()
	rows, err := db.Query("select task_id, group_id, task, completed, created_at, completed_at from servtasks")
	CheckError(err)
	var CreatedAt string
	var CompletedAt string
	for rows.Next(){
		err := rows.Scan(&t.TaskId, &t.GroupId, &t.Task, &t.Completed, &CreatedAt, &CompletedAt)
		CheckError(err)
		t.CreatedAt, err = time.Parse("2006-01-02 15:04:05.0+00", CreatedAt)
		CheckError(err)
		t.CompletedAt, err = time.Parse("2006-01-02 15:04:05.0+00", CompletedAt)
		CheckError(err)
		Tasks = append(Tasks, t)
	}
	defer rows.Close()
	return Tasks
}

// Отправка клиенту задач по Echo
func SendingToKlientTaskforEcho(c echo.Context, Tasks[] Task, limit int) {
	for i := 0; i < len(Tasks) && i < limit; i++ {
		b, _ := json.Marshal(Tasks[i])
		c.String(http.StatusOK, strconv.Itoa(i) + ") " + string(b) + "\n\n")
	}
}

// Сортировка задач по имени
func SortTaskByNamePlus(Tasks[] Task, typetask string) []Task{
	Tasks = SortTaskByName(Tasks)
	if typetask == "" || typetask == "all" {
		return Tasks
	} else {
		Tasks = OnlyNeedType(Tasks, typetask)
		return Tasks
	}
}

// Отбираем только задачи с нужным типом
func OnlyNeedType(Tasks[] Task, typetask string) []Task{
	var TypedTask[] Task
	for _, n := range Tasks{
		if typetask == "completed" {
			if n.Completed {
				TypedTask = append(TypedTask, n)
			}
		} else {
			if !n.Completed {
				TypedTask = append(TypedTask, n)
			}
		}
	}
	return TypedTask
}

// Сортировка задач по группам и по типу
func SortTaskByGroupPlus(Tasks[] Task, typetask string) []Task{
	Tasks = SortTaskByGroup(Tasks)
	if typetask == "" || typetask == "all" {
		return Tasks
	} else {
		Tasks = OnlyNeedType(Tasks, typetask)
		return Tasks
	}
}

// Проверка id элемента в списке задач
func ContainsTaskId (list []Task, x string) bool {
	for _, n := range list {
		if x == n.TaskId {
			return true
		}
	}
	return false
}

// Обновление данных о задачи в БД
func UpdateTaskbyId(Task Task, id string){
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	sqlStatement := `UPDATE servtasks SET task_id = $1, group_id = $2, task = $3 WHERE task_id = $4;`
	str := string(Task.GroupId) + Task.Task
	str = HashT(str)
	Task.TaskId = str[:5]
	_, err = db.Exec(sqlStatement, Task.TaskId, Task.GroupId, Task.Task, id)
	CheckError(err)
}

// Обновление статуса задачи в БД
func UpdateStatusbyId(Task Task, id string, Time time.Time){
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	completedAt := FormatDate(Time)
	sqlStatement := `UPDATE servtasks SET completed = $1, completed_at = $2 WHERE task_id = $3;`
	_, err = db.Exec(sqlStatement, Task.Completed, completedAt, id)
	CheckError(err)
}

// Получение задачи по ID
func FindTaskById(id string, Tasks[] Task) Task{
	for _, n := range Tasks {
		if n.TaskId == id {
			return n
		}
	}
	var s Task
	s.TaskId = "-1"
	return s
}

// Получение задач по группам
func GetTasksByGroup(Tasks[] Task, GroupId int, typeComplete string) []Task{
	var TaskGroup[] Task
	for _,n := range Tasks {
		if n.GroupId == GroupId{
			if typeComplete == "completed" && n.Completed{
				TaskGroup = append(TaskGroup, n)
			} else {
				if typeComplete == "working" && !n.Completed {
					TaskGroup = append(TaskGroup, n)
				} else {
					if typeComplete == "all" || typeComplete == "" {
						TaskGroup = append(TaskGroup, n)
					}
				}
			}
		}
	}
	return TaskGroup
}

// Получение из строки булево значение
func StringToBool(str string) (bool, error) {
	str = strings.ToLower(str)
	if str == "true" {
		return true, nil
	} else {
		if str == "false" {
			return false, nil
		} else {
			return false, errors.New("Wrong string")
		}
	}
}

func Statistica(Period time.Duration, Tasks[] Task) (int, int){
	t := time.Now().Add(-Period)
	var completed, created int
	for _, n := range Tasks {
		if n.CreatedAt.After(t) {
			created++
		}
		if n.CompletedAt.After(t) {
			completed++
		}
	}
	return completed, created
}

func Yesterday(Tasks[] Task) (int, int){
	t := time.Now().Add(-time.Hour*24*2)
	var completed, created int
	for _, n := range Tasks {
		if n.CreatedAt.After(t) && n.CreatedAt.Before(time.Now().Add(-time.Hour*24)) {
			created++
		}
		if n.CompletedAt.After(t) && n.CompletedAt.Before(time.Now().Add(-time.Hour*24)){
			completed++
		}
	}
	return completed, created
}