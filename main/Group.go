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
)

// Валидация данных
func (g Group)  ValidateGroup() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.GroupName, validation.Required, validation.Length(1, 50)),
		validation.Field(&g.GroupDescription, validation.Length(0, 150)),
		)
}

// Логирование группы
func (g Group) LogingGroup(){
	log.WithFields(log.Fields{
		"group_id": g.GroupId,
		"group_name": g.GroupName,
		"group_description": g.GroupDescription,
		"parent_id": g.ParentId,
	}).Info("")
}

// /group/new Обработка запроса создания новой группы
func HandlePostNewGroupEcho(c echo.Context) error{
	log.Println("Begin create new group")
	var u Group
	body, err := ioutil.ReadAll(c.Request().Body)
	CheckError(err)
	json.Unmarshal(body, &u)
	err = u.ValidateGroup()
	if err == nil{
		if !IsExistByName(u.GroupName, GetGroupsFromBD()){
			GroupInBD(u.GroupName, u.GroupDescription, u.ParentId)
			b, _ := json.Marshal(u)
			u.LogingGroup()
			log.Println("Group create successfully")
			return c.String(http.StatusCreated, string(b))
		} else {
			u.LogingGroup()
			log.Println("This Group name is taken")
			return c.String(http.StatusAccepted, "This Group name is taken")
		}
	} else {
		CheckError(err)
		u.LogingGroup()
		log.Println("Crash create group")
		return c.String(http.StatusBadRequest, err.Error())
	}
}

// /group Обработка запроса получения групп
func HandleGetGroupEcho(c echo.Context) error{
	log.Println("Begin sort group")
	filter := c.QueryParam("sort")
	var limit int
	if c.QueryParam("limit") == "" {
		limit = 9223372036854775807
	} else {
		limit, _ = strconv.Atoi(c.QueryParam("limit"))
	}
	switch filter {
	case "name":
		Gr := SortGroupByName(GetGroupsFromBD())
		SendingToKlientGroupforEcho(c, Gr, limit)

	case "parents_first":
		Gr := SortParentFirst(GetGroupsFromBD())
		SendingToKlientGroupforEcho(c, Gr, limit)

	case "parent_with_childs":
		Gr := SortParentWithChild(GetGroupsFromBD())
		SendingToKlientGroupforEcho(c, Gr, limit)
	default:
		Gr := GetGroupsFromBD()
		SendingToKlientGroupforEcho(c, Gr, limit)
	}
	log.Println("Sort completed")
	return c.String(http.StatusOK, "")
}

// Получение группы по ID
func GetGroupById (c echo.Context) error {
	log.Println("Begin found group")
	i := c.Param("id")
	id, _ := strconv.Atoi(i)
	Group := FindGroupById(id, GetGroupsFromBD())
	if Group.GroupId != -1 {
		log.Println("Group found successfully")
		return c.String(http.StatusOK, fmt.Sprintf("Group ID:%d\nGroup name:%s\nGroup parentID:%d\nGroup description:%s",Group.GroupId, Group.GroupName, Group.ParentId, Group.GroupDescription))
	} else {
		log.Println("Group not found")
		return c.String(http.StatusAccepted, "Group with this ID not found")
	}
}

// ПОлучение списка детей группы по ID
func GetChildById(c echo.Context) error {
	log.Println("Begin found child")
	i := c.Param("id")
	id, _ := strconv.Atoi(i)
	Child := ChildsById(GetGroupsFromBD(), id)
	var OnSend string
	for _, n := range Child {
		OnSend += fmt.Sprintf("Group ID:%d\nGroup name:%s\nGroup parentID:%d\nGroup description:%s\n\n", n.GroupId, n.GroupName, n.ParentId, n.GroupDescription)
	}
	if OnSend != "" {
		log.Println("Child found successfully")
		return c.String(http.StatusOK, OnSend)
	} else {
		log.Println("Child not found")
		return c.String(http.StatusAccepted, "Child ton Found")
	}
}

// Удаление группы по ID
func DeleteById(c echo.Context) error {
	log.Println("Begin delete group")
	i := c.Param("id")
	id, _ := strconv.Atoi(i)
	g := FindGroupById(id, GetGroupsFromBD())
	log.Println("Delete this group")
	g.LogingGroup()
	err := DeleteGroupFromBDbyId(id)
	if err == 0 {
		log.Println("Group deleted successfully")
	return c.String(http.StatusOK, "item deleted")
	} else {
		log.Println("Crushed deleted. May be group has child groups or nested tasks")
	return c.String(http.StatusAccepted, "The group has child groups or nested tasks")
	}
}

// Обновить данные группы по ID
func PutGroupbyId(c echo.Context) error{
	log.Println("Begin change group data")
	i := c.Param("id")
	id, _ := strconv.Atoi(i)
	body, err := ioutil.ReadAll(c.Request().Body)
	var u Group
	json.Unmarshal(body, &u)
	err = u.ValidateGroup()
	if err == nil {
		fmt.Println("Group", u.GroupName)
		CheckError(err)
		UpdateGroupbyId(u, id)
		u = FindGroupById(id, GetGroupsFromBD())
		b, _ := json.Marshal(u)
		if u.GroupId != -1 {
			log.Println("Group successfully change")
			u.LogingGroup()
			return c.String(http.StatusOK, string(b))
		} else {
			log.Println("Change group crashed. Group with this ID not found")
			return c.String(http.StatusAccepted, "Group with this ID not found")
		}
	} else {
		log.Println("Change group crashed. Invalid data")
		u.LogingGroup()
		return c.String(http.StatusBadRequest, err.Error())
	}
}

