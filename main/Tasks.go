package main

import (
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (t Task) ValidateTask() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Task, validation.Required, validation.Length(3, 200)))
}

// Обработка запроса на получение задачи (/task)
func GetTask(c echo.Context) error{
	log.Println("Start sort task")
	filter := c.QueryParam("sort")
	var limit int
	if c.QueryParam("limit") == "" {
		limit = 9223372036854775807
	} else {
		limit, _ = strconv.Atoi(c.QueryParam("limit"))
	}
	typetask := c.QueryParam("type")
	switch filter {
	case "name":
		Ts := SortTaskByNamePlus(GetTasksFromBD(), typetask)
		SendingToKlientTaskforEcho(c, Ts, limit)

	case "groups":
		Ts := SortTaskByGroupPlus(GetTasksFromBD(), typetask)
		SendingToKlientTaskforEcho(c, Ts, limit)

	default:
		Ts := GetTasksFromBD()
		SendingToKlientTaskforEcho(c, Ts, limit)
	}
	log.Println("Task successfully sort")
	return c.String(http.StatusOK, "")
}

// Обработка запроса на создание новой задачи (/task/new)
func CreateNewTask(c echo.Context) error{
	log.Println("Start create new task")
	var t Task
	body, err := ioutil.ReadAll(c.Request().Body)
	CheckError(err)
	json.Unmarshal(body, &t)
	err = t.ValidateTask()
	if err == nil {
		str := string(t.GroupId) + t.Task
		str = HashT(str)
		t.TaskId = str[:5]
		t.CreatedAt = time.Now()
		t.Completed = false
		if !ContainsTaskId(GetTasksFromBD(), t.TaskId[:5]){
			TaskToBD(t)
			fmt.Println("Connected!")
			fmt.Println(string(body))
			b, _ := json.Marshal(t)
			log.Println("Task successfully created")
			t.LogingTask()
			return c.String(http.StatusCreated, string(b) + "\n")
		} else {
			log.Println("Create new task crashed. Element already exists")
			return c.String(http.StatusAccepted, "Element already exists")
		}
	} else {
		log.Println("Create new task crashed. Invalid data")
		CheckError(err)
		t.LogingTask()
		return c.String(http.StatusBadRequest, err.Error())
	}
}

// Обновить данные задачи по ID (/task/:id)
func PutTaskbyId(c echo.Context) error{
	log.Println("Begin change task")
	id := c.Param("id")
	body, err := ioutil.ReadAll(c.Request().Body)
	CheckError(err)
	var t Task
	json.Unmarshal(body, &t)
	if ContainsTaskId(GetTasksFromBD(), id) {
		t.TaskId = id
		UpdateTaskbyId(t, id)
		b, _ := json.Marshal(t)
		log.Println("Task successfully changed")
		return c.JSON(http.StatusCreated, string(b))
	} else {
		log.Println("Changed failed. Task with this ID not found")
		return c.String(http.StatusAccepted, "Task with this ID not found")
	}
}

// Обработка запроса получения заданий группы по типу (/task/group/:id)
func GetTasksbyType(c echo.Context) error{
	log.Println("Start get task by type")
	id := c.Param("id")
	IdGroup, err := strconv.Atoi(id)
	type_complete := c.QueryParam("type")
	CheckError(err)
	SendingToKlientTaskforEcho(c,GetTasksByGroup(GetTasksFromBD(), IdGroup, type_complete), 9223372036854775807)
	return c.String(http.StatusOK, "")
}

// Обработка запроса изменения статуса завершенности задачи (/taskm/:id)
func CompletingTask(c echo.Context) error{
	log.Println("Start changed status")
	id := c.Param("id")
	typeCompleting := c.QueryParam("finished")
	typeComplet, err := StringToBool(typeCompleting)
	CheckError(err)
	if err == nil {
		var Time time.Time
		s := FindTaskById(id, GetTasksFromBD())
		if typeComplet != s.Completed {
			s.Completed = typeComplet
			if typeComplet {
				Time = time.Now()
			}
			UpdateStatusbyId(s, id, Time)
			s := FindTaskById(id, GetTasksFromBD())
			b, _ := json.Marshal(s)
			log.Println("Changed status successfully completed")
			s.LogingTask()
			return c.String(http.StatusOK, string(b))
		} else {
			log.Println("Changed status failed. Status already this")
			return c.String(http.StatusAccepted, "Status already this")
		}
	} else {
		log.Println("Changed status failed. Parameter not can be bool")
		CheckError(err)
		return c.String(http.StatusBadRequest, err.Error())
	}
}

// Обработка запроса получения статистики (/stat/:period)
func Statistic(c echo.Context) error{
	log.Println("Start calculate statistic")
	period := c.Param("period")
	var completed, created int
	switch period {
	case "today":
		completed, created = Statistica(time.Hour*24, GetTasksFromBD())
	case "yesterday":
		completed, created = Yesterday(GetTasksFromBD())
	case "week":
		completed, created = Statistica(time.Hour*24*7, GetTasksFromBD())
	case "month":
		completed, created = Statistica(time.Hour*24*30, GetTasksFromBD())
	}
	log.Println("Statistic successfully collected")
	return c.String(http.StatusOK, fmt.Sprintf("completed: %d\ncreated: %d", completed, created))
}

func (t Task) LogingTask(){
	log.WithFields(log.Fields{
		"task_id": t.TaskId,
		"group_id": t.GroupId,
		"task": t.Task,
		"completed": t.Completed,
		"created_at": t.CreatedAt,
		"completed_at": t.CompletedAt,
	}).Info("")
}
