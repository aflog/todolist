package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var a = App{}
var testConfig = Config{
	AppName:     "todolist",
	MysqlUser:   "root",
	MysqlPwd:    "testroot",
	MysqlHost:   "127.0.0.1",
	MysqlPort:   "3307",
	MysqlBDName: "todolist",
}

func TestMain(m *testing.M) {

	if err := a.Initialize(testConfig); err != nil {
		log.Fatal(err)
	}
	ensureTableExists()
	code := m.Run()
	clearTable()
	defer a.db.Close()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/items", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	body := response.Body.String()
	if strings.Trim(body, "\n") != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentItem(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/items/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	body := response.Body.String()
	if strings.Trim(body, "\n") != "Item not found." {
		t.Errorf("Expected the response 'Item not found'. Got '%s'", body)
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(testItemJSON1)

	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	body := response.Body.String()
	if strings.Trim(body, "\n") != `{"id":1}` {
		t.Errorf(`Expected to receive id 1 in the format '{"id":1}'. Got '%s'`, body)
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()

	// add item to the db
	addItems(1)
	addComments(2, 1)
	addLabels(1, 1)

	// prepare the expected Json for comparism, the initial struct is used for readability.
	expectedData := expectedStruct{
		ID:    1,
		Title: "Test automatic title 1",
		Comments: []commentStruct{
			{ID: 1, Text: "Test automatic comment 1"},
			{ID: 2, Text: "Test automatic comment 2"},
		},
		Labels: []labelStruct{
			{ID: 1, Text: "Test automatic label 1"},
		},
		Description: "Test automatic description 1",
		DueDate:     "2021-05-15T13:11:50Z",
		Status:      false,
	}
	expectedJSON, err := json.Marshal(expectedData)
	if err != nil {
		log.Fatal(err.Error())
	}

	// request the created item
	req, _ := http.NewRequest("GET", "/items/1", nil)
	response := executeRequest(req)

	// check the response
	checkResponseCode(t, http.StatusOK, response.Code)

	body := response.Body.String()
	expectedString := bytes.NewBuffer(expectedJSON).String()
	if strings.Trim(body, "\n") != expectedString {
		t.Errorf(`Expected to receive id 1 in the format '%s'. Got '%s'`, expectedString, body)
	}
}

func TestNotFoundRoot(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func addItems(count int) {
	layout := "2006-01-02T15:04:05.9Z"
	due, err := time.Parse(layout, "2021-05-15T13:11:50Z")
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= count; i++ {
		_, err := a.db.Exec("INSERT INTO item(title, description, due) VALUES(?, ?, ?)", fmt.Sprintf("Test automatic title %d", i), fmt.Sprintf("Test automatic description %d", i), due)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func addComments(count int, itemID int) {
	for i := 1; i <= count; i++ {
		_, err := a.db.Exec("INSERT INTO comment(itemId, comment) VALUES(?, ?)", itemID, fmt.Sprintf("Test automatic comment %d", i))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func addLabels(count int, itemID int) {
	for i := 1; i <= count; i++ {
		_, err := a.db.Exec("INSERT INTO label(itemId, label) VALUES(?, ?)", itemID, fmt.Sprintf("Test automatic label %d", i))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)

	return rr
}

func ensureTableExists() {
	if _, err := a.db.Exec(tableItemCreationQuery); err != nil {
		fmt.Println("fail to execute")
		log.Fatal(err)
	}
	if _, err := a.db.Exec(tableCommentCreationQuery); err != nil {
		fmt.Println("fail to execute")
		log.Fatal(err)
	}
	if _, err := a.db.Exec(tableLabelCreationQuery); err != nil {
		fmt.Println("fail to execute")
		log.Fatal(err)
	}
}

func clearTable() {
	a.db.Exec("DELETE FROM todolist.item")
	a.db.Exec("ALTER TABLE todolist.item AUTO_INCREMENT = 1")
	a.db.Exec("DELETE FROM todolist.comment")
	a.db.Exec("ALTER TABLE todolist.comment AUTO_INCREMENT = 1")
	a.db.Exec("DELETE FROM todolist.label")
	a.db.Exec("ALTER TABLE todolist.label AUTO_INCREMENT = 1")
}

type expectedStruct struct {
	ID          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Labels      []labelStruct   `json:"labels"`
	Comments    []commentStruct `json:"comments"`
	Status      bool            `json:"status"`
	DueDate     string          `json:"dueDate"`
}

type commentStruct struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type labelStruct struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

const testItemJSON1 = `{
	"title":"test title 1",
	"description":"test decription 1",
	"dueDate":"2021-05-15T13:11:50Z",
	"comments":[{"text":"test comment 1"},{"text":"test comment 2"}],
	"labels":[{"text":"test label 1"},{"text":"test label 2"}]
	}`

const tableItemCreationQuery = `CREATE TABLE IF NOT EXISTS todolist.item (
	id INT(6) NOT NULL AUTO_INCREMENT,
	title VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
	description VARCHAR(500) CHARACTER SET utf8 COLLATE utf8_unicode_ci,
	status BOOLEAN NOT NULL DEFAULT false,
	due DATETIME,
	created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;`

const tableCommentCreationQuery = `CREATE TABLE IF NOT EXISTS todolist.comment (
    id INT(6) NOT NULL AUTO_INCREMENT,
    itemId INT(6) NOT NULL,
    comment VARCHAR(500) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (itemId) 
        REFERENCES item(id) 
        ON DELETE CASCADE,
    INDEX (itemId)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;`

const tableLabelCreationQuery = `CREATE TABLE IF NOT EXISTS todolist.label (
    id INT(6) NOT NULL AUTO_INCREMENT,
    itemId INT(6) NOT NULL,
    label VARCHAR(500) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (itemId)
        REFERENCES item(id)
        ON DELETE CASCADE,
    INDEX (itemId)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;`
