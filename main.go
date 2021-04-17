package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	_ "github.com/mattn/go-sqlite3"
)

type DataUsers struct {
	Response struct {
		Count int64   `json:"count"`
		Items []int64 `json:"items"`
	} `json:"response"`
}

type SettingStruct struct {
	name  string
	token string
}

type UserStruct struct {
	id       int
	user_id  string
	group_id string
	status   int
}

type RseponsAddFrienStructure struct {
	Response int `json:"response"`
}

func main() {

	var isSetting int
	var db *sql.DB
	var name string
	var token string
	db = InitDB()
	isSetting = ReadItem(db)
	if isSetting == 0 {
		CreateTable(db)
		fmt.Print("Введите имя: ")
		fmt.Scan(&name)
		fmt.Print("Введите токен: ")
		fmt.Scan(&token)
		insertSql(db, "INSERT INTO setting(name, token) values('"+name+"','"+token+"')")
	}
	if os.Args[1] == "group" {
		if err := root(db, os.Args[1:2]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if os.Args[1] == "add_friend" {
		if len(os.Args) == 3 {
			arg, _ := strconv.ParseUint(os.Args[2], 10, 64)
			gocron.Every(arg).Minutes().Do(rootChild)
		} else {
			gocron.Every(47).Minutes().Do(rootChild)
		}
		<-gocron.Start()
	}

	//getSql()
	//go getRequest()

	fmt.Print("Aloha")
}

func rootChild() {
	var db *sql.DB
	db = InitDB()
	str := make([]string, 1, 2)
	str[0] = "add_friend"
	root(db, str)
}

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./my.db")
	if err != nil {
		panic(err)
	}
	/* stmt, err := db.Prepare("INSERT INTO groups(group_id) values(?)")
	if err != nil {
		panic(err)
	}
	res, err := stmt.Exec("255599")
	if err != nil {
		panic(err)
	} */
	return db

}

func insertSql(db *sql.DB, query string) {

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

}

func selectSql(db *sql.DB, query string) *sql.Rows {

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	return rows
}

func ReadItem(db *sql.DB) int {
	sql_readall := `
	SELECT name FROM sqlite_master WHERE type='table' AND name='setting';
	`
	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var result []SettingStruct
	for rows.Next() {
		item := SettingStruct{}
		err2 := rows.Scan(&item.name)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return len(result)
}

func root(db *sql.DB, args []string) error {
	if len(args) < 1 {
		return errors.New("You must pass a sub-command")
	}

	if args[0] == "add_friend" {
		var token string
		token = getToken(db)
		adf := addFriend(db, token)
		upsql(db, adf[0], adf[1])
		return errors.New("11111")
	}

	if os.Args[1] == "group" {

		addPeoples(db, os.Args[2])
		return errors.New("11111")
	}
	if os.Args[1] == "add_friend" {
		var token string
		token = getToken(db)
		adf := addFriend(db, token)

		upsql(db, adf[0], adf[1])
		return errors.New("11111")
	}
	return errors.New("54545")

}

func addFriend(db *sql.DB, token string) []string {
	var data []byte
	rows := selectSql(db, "SELECT user_id FROM users WHERE status = '0' ORDER BY user_id DESC")
	defer rows.Close()
	rows.Next()
	itemData := UserStruct{}
	err := rows.Scan(&itemData.user_id)
	if err != nil {
		fmt.Println(err)
	}

	res, err := http.Get("https://api.vk.com/method/friends.add?user_id=" + itemData.user_id + "&v=5.130&access_token=" + token)

	// check for response error
	if err != nil {
		log.Fatal(err)
	}

	// read all response body
	data, err = ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}
	var item RseponsAddFrienStructure
	//fmt.Println(data)
	//datas, err := json.Marshal(data)

	json.Unmarshal(data, &item)
	arr := make([]string, 2)

	arr[0] = strconv.Itoa(item.Response)
	arr[1] = itemData.user_id
	return arr

}

func upsql(db *sql.DB, iresponse string, userid string) {
	if iresponse == "1" {
		insertSql(db, "UPDATE users SET status = '1' WHERE user_id = '"+userid+"'") // success send
	} else if iresponse == "0" {
		insertSql(db, "UPDATE users SET status = '2' WHERE user_id = '"+userid+"'") // user have block into vk.com
	} else {
		fmt.Println("error")
	}
}

func getToken(db *sql.DB) string {
	rows := selectSql(db, "SELECT token FROM setting WHERE name = 'Мастеров' LIMIT 1")
	defer rows.Close()
	rows.Next()
	item := SettingStruct{}
	rows.Scan(&item.token)
	return item.token
}

func addPeoples(db *sql.DB, idGroup string) {
	var result [2]string
	rows := selectSql(db, "SELECT * FROM setting WHERE name = 'Мастеров'")
	defer rows.Close()

	for rows.Next() {
		item := SettingStruct{}
		err := rows.Scan(&item.name, &item.token)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result[0] = item.name
		result[1] = item.token
	}
	uDataId := getRequest(result[0], idGroup)

	var dataStr string
	for _, v := range uDataId {
		strV := fmt.Sprint(v)
		dataStr += "('" + strV + "'),"
	}

	// close response body
	dataStr = strings.TrimSuffix(dataStr, ",")
	insertSql(db, "INSERT INTO users(user_id) values "+dataStr)
}

func CreateTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE "groups" (
		"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"group_id"	INTEGER UNIQUE,
		"status"	INTEGER
	);
	CREATE TABLE "setting" (
		"token"	INTEGER NOT NULL UNIQUE,
		"name"	INTEGER NOT NULL UNIQUE
	);
	CREATE TABLE "users" (
		"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"user_id"	INTEGER,
		"group_id"	INTEGER,
		"status"	INTEGER DEFAULT 0
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func getRequest(token string, idGroup string) []interface{} {
	// make a sample HTTP GET request
	var data []byte
	var countUsers int
	var offsetUsers int
	var dataUsers []interface{}

	res, err := http.Get("https://api.vk.com/method/groups.getMembers?group_id=" + idGroup + "&count=1000&offset=0&v=5.130&access_token=" + token)

	// check for response error
	if err != nil {
		log.Fatal(err)
	}

	// read all response body
	data, err = ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}
	var item DataUsers
	//fmt.Println(data)
	//datas, err := json.Marshal(data)
	json.Unmarshal(data, &item)

	countUsers = int(item.Response.Count) / 1000

	for i := 0; i < countUsers; i++ {
		time.Sleep(3 * time.Second)
		offsetUsers = i * 1000
		res, err := http.Get("https://api.vk.com/method/groups.getMembers?group_id=" + idGroup + "&count=1000&offset=" + strconv.Itoa(offsetUsers) + "&v=5.130&access_token=" + token)

		// check for response error
		if err != nil {
			log.Fatal(err)
		}

		// read all response body
		data, err = ioutil.ReadAll(res.Body)

		if err != nil {
			fmt.Println(err)
		}
		var item DataUsers

		json.Unmarshal(data, &item)
		for _, v := range item.Response.Items {
			dataUsers = append(dataUsers, v)
		}

	}

	return dataUsers

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
