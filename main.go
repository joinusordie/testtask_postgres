package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/joinusordie/testtask_postgres/repository"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var watcher *fsnotify.Watcher

func main() {

	temp := make([]string, 0)

	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env variables: %s", err.Error())
	}
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository.NewRepository(db)

	// Create new watcher.
	watcher, _ = fsnotify.NewWatcher()

	if err := filepath.Walk(viper.GetString("path"), watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	logrus.Print("Observer Started")

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				repos.RecordLog(event)
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
				for index, value := range viper.GetStringSlice("commands") {
					temp = strings.Split(value, " ")
					cmd := exec.Command(temp[0], temp[1:]...)
					cmd.Dir = viper.GetString("path")
					fmt.Println(cmd)
					err := cmd.Run()
					if err != nil {
						fmt.Printf("Команда №%d в путе %s не выполнилась, пропускаю остальные", index+1, viper.GetString("path"))
						break
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Observer Shutting Down")

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connectrion close: %s", err.Error())
	}

	if err := watcher.Close(); err != nil {
		logrus.Errorf("error occured on observer close: %s", err.Error())
	}
}

func watchDir(path string, fi os.FileInfo, err error) error {

	if fi.IsDir() && fi.Name() == "build" {
		return filepath.SkipDir
	}
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
