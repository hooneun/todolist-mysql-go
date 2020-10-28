package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/cors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:4444)/todolist?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})

// TodoItem !
type TodoItem struct {
	ID          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

// Healthz !
func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

// CreateItem !
func CreateItem(w http.ResponseWriter, r *http.Request) {
	desc := r.FormValue("description")
	log.WithFields(log.Fields{"description": desc}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItem{Description: desc, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)

	if result.Error != nil {
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(&todo)
}

// UpdateItem !
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	err := GetItemByID(id)
	if !err {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
		return
	}

	completed, _ := strconv.ParseBool(r.FormValue("completed"))
	log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating TodoItem")
	todo := &TodoItem{}
	db.First(&todo, id)
	todo.Completed = completed
	db.Save(&todo)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"updated": true}`)
}

// DeleteItem !
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItemByID(id)
	if !err {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record Not Found"}`)
		return
	}

	log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
	todo := &TodoItem{}
	db.First(&todo, id)
	db.Delete(&todo)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"deleted": true}`)
}

// GetItemByID !
func GetItemByID(ID int) bool {
	todo := &TodoItem{}
	result := db.First(&todo, ID)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

// GetCompletedItems !
func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Completed TodoItems")
	completedItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedItems)
}

// GetIncompleteItems !
func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Incomplete TodoItems")
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)
}

// GetTodoItems !
func GetTodoItems(completed bool) interface{} {
	var todos []TodoItem
	todoItems := db.Where("completed = ?", completed).Find(&todos)

	if todoItems.Error != nil {

	}
	return todos
}

func main() {
	log.Info("Starting Todolist API server")

	// db.Debug().Migrator().DropTable(&TodoItem{})
	// db.Debug().Migrator().AutoMigrate(&TodoItem{})

	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	http.ListenAndServe(":8000", handler)
}
