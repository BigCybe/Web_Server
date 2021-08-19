package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"sort"
	"strconv"
)

type GroupByName []Group

func (a GroupByName) Len() int           { return len(a) }
func (a GroupByName) Less(i, j int) bool { return a[i].GroupName < a[j].GroupName }
func (a GroupByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

//Добавление группы в базу данных
func GroupInBD(Name, Description string, id int){
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()
	insertStmt := `insert into  "servgroups"("group_name", "group_description", "parent_id") values($1, $2, $3)` //Запись данных в БД
	_, e := db.Exec(insertStmt, Name, Description, id)
	CheckError(e)
	err = db.Ping()
	CheckError(err)
}

// Получение списка групп из БД
func GetGroupsFromBD() []Group{
	var u Group
	var Groups[] Group
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()
	rows, err := db.Query("select group_name, group_description, group_id, parent_id from servgroups")
	CheckError(err)
	for rows.Next(){
		err := rows.Scan(&u.GroupName, &u.GroupDescription, &u.GroupId, &u.ParentId)
		CheckError(err)
		Groups = append(Groups, u)
	}
	defer rows.Close()
	return Groups
}

// Сортировка групп по имени
func SortGroupByName(Groups[] Group)  []Group {
	sort.Sort(GroupByName(Groups))
	return Groups
}

// Сортировка где сперва родители
func SortParentFirst(Groups[] Group) []Group{
	var ParentId[] int
	var ParentFirst[] Group
	var Child[] Group
	ParentId = AllIdParent(Groups)
	for _, n := range Groups {
		if ContainsInt(ParentId, n.GroupId){
			ParentFirst = append(ParentFirst, n)
		} else {
			Child = append(Child, n)
		}
	}
	ParentFirst = SortGroupByName(ParentFirst)
	Child = SortGroupByName(Child)
	ParentFirst = AppendMasbyMas(ParentFirst, Child)
	return ParentFirst
}

// Сортировка по типу Родитель - Дети
func SortParentWithChild(Groups[] Group) []Group {
	var ParentId[] int
	var Parents[] Group
	var ParentChild[] Group
	var GroupChild[] Group
	ParentId = AllIdParent(Groups)
	for _, n := range Groups { // Составляем группу родителей
		if ContainsInt(ParentId, n.GroupId) {
			Parents = append(Parents, n)
		}
	}
	Parents = SortGroupByName(Parents)

	for _, n := range Parents {
		ParentChild = append(ParentChild, n)
		for _, c := range Groups {
			if c.ParentId == n.GroupId {
				GroupChild = append(GroupChild, c)
			}
		}
		GroupChild = SortGroupByName(GroupChild)
		ParentChild = AppendMasbyMas(ParentChild, GroupChild)
		GroupChild = nil
	}

	return ParentChild
}

// Добавление массива 2 к массиву 1
func AppendMasbyMas (mas1[] Group, mas2[] Group) []Group{
	for _, n := range mas2 {
		mas1 = append(mas1, n)
	}
	return mas1
}

// Определение всех id родителей группы
func AllIdParent(Groups[] Group) []int{
	var ParentId[] int

	for _, n := range Groups {
		if string(n.ParentId) != "" && !ContainsInt(ParentId, n.ParentId) {
			ParentId = append(ParentId, n.ParentId)
		}
	}

	return ParentId
}

// Проверка наличия элемента в массиве (int)
func ContainsInt (list []int, x int) bool {
	for _, n := range list {
		if x == n {
			return true
		}
	}
	return false
}

// Отправка клиенту групп
func SendingToKlient (w Writer, Groups[] Group, limit int){
	for i := 0; i < len(Groups) && i < limit; i++ {
		b, _ := json.Marshal(Groups[i])
		io.WriteString(w, string(b)+"\n")
	}
}
// Отправка клиенту групп по Echo
func SendingToKlientGroupforEcho(c echo.Context, Groups[] Group, limit int) {
	for i := 0; i < len(Groups) && i < limit; i++ {
		b, _ := json.Marshal(Groups[i])
		c.String(http.StatusOK, strconv.Itoa(i) + ") " + string(b) + "\n\n")
	}
}

// Получение группы по ID
func FindGroupById(id int, Groups[] Group) (Group){
	for _, n := range Groups {
		if n.GroupId == id {
			return n
		}
	}
	var s Group
	s.GroupId = -1
	return s
}

func IsExistByName(Name string, Groups[] Group) bool {
	IsExist := false
	for _, n := range Groups {
		if n.GroupName == Name {
			IsExist = true
		}
	}
	return IsExist
}

// Удаление группы из БД по ID
func DeleteGroupFromBDbyId(id int) int {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	CheckError(err)
	defer db.Close()
	if !ContainsInt(AllIdParent(GetGroupsFromBD()), id) {
		result, err := db.Exec("delete from servgroups where group_id = $1", id)
		CheckError(err)
		fmt.Println(result.RowsAffected())
		return 0
	} else {
		return 1
	}
}

// Получение списка детей группы
func ChildsById(Groups[] Group, id int) []Group{
	var Child[] Group
	for _, n := range Groups {
		if n.ParentId == id {
			Child = append(Child, n)
		}
	}
	return Child
}

// Обновление данных о группе в БД
func UpdateGroupbyId(Group Group, id int){
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open(user, psqlconn)
	sqlStatement := `UPDATE servgroups SET group_name = $1, group_description = $2, parent_id = $3 WHERE group_id = $4;`
	_, err = db.Exec(sqlStatement, Group.GroupName, Group.GroupDescription, Group.ParentId, id)
	CheckError(err)
}
