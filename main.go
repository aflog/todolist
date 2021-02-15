package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aflog/todolist/handler"
	"github.com/aflog/todolist/repository/mysql"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// load configuration variables
	conf, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// set and run the application
	app := App{}
	if err = app.Initialize(conf); err != nil {
		log.Fatal(err)
	}
	defer app.db.Close()

	app.Run(":8000")
}

//Health handler sends an alive response
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // Set 200 OK
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`) // Send json to the ResponseWriter
}

// Config data of the application
type Config struct {
	AppName     string `mapstructure:"APP_NAME"`
	MysqlUser   string `mapstructure:"MYSQL_APP_USER"`
	MysqlPwd    string `mapstructure:"MYSQL_APP_PASSWORD"`
	MysqlHost   string `mapstructure:"MYSQL_HOST"`
	MysqlPort   string `mapstructure:"MYSQL_PORT"`
	MysqlBDName string `mapstructure:"MYSQL_DB_NAME"`
}

// LoadConfig creates the configuration from flags, env and file.
// The precedence order from highest to lowest is flag->env->file.
// Environment variables should be prefixed with TODOLIST_
func LoadConfig() (c Config, err error) {
	// TODO implement the .env file -flag and load it to the config
	/*flag.String("conf", "", "path to configuration .env file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)*/

	viper.SetEnvPrefix("TODOLIST")
	viper.AutomaticEnv()

	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)
	return
}

// App allows to set up and run the application
type App struct {
	conf   Config
	db     *sql.DB
	router *mux.Router
}

// Initialize sets up the application
func (a *App) Initialize(c Config) error {
	a.conf = c
	var err error
	// initialize the DB
	a.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?parseTime=true", a.conf.MysqlUser, a.conf.MysqlPwd, a.conf.MysqlHost, a.conf.MysqlPort, a.conf.MysqlBDName))
	if err != nil {
		return err
	}
	err = a.db.Ping()
	if err != nil {
		return err
	}

	// get repository for list items
	sqlRepo := mysql.NewRepository(a.db)
	itemsHandler, err := handler.New(sqlRepo)
	if err != nil {
		return err
	}

	log.Println("Starting Todolist API server")
	a.router = mux.NewRouter()
	a.router.HandleFunc("/health", Health).Methods("GET")
	// TODO router.HandleFunc("/items/{id}/done", ToggleDone).Methods("POST")
	a.router.HandleFunc("/items/{id}", itemsHandler.Select).Methods("GET")
	// TODO router.HandleFunc("/items/{id}", itemsHandler.Update).Methods("UPDATE")
	a.router.HandleFunc("/items", itemsHandler.List).Methods("GET")
	a.router.HandleFunc("/items", itemsHandler.Add).Methods("POST")

	return nil
}

// Run starts the application
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.router))
}
