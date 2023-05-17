package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
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

type ObserverItem struct {
	Path          string   `mapstructure:"path"`
	Commands      []string `mapstructure:"commands"`
	IncludeRegexp []string `mapstructure:"include_regexp"`
	ExcludeRegexp []string `mapstructure:"exclude_regexp"`
	LogFile       string   `mapstructure:"log_file"`
}

type Config struct {
	Observer []ObserverItem `mapstructure:"obs"`
}

func main() {

	var conf *Config

	temp := make([]string, 0)

	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env variables: %s", err.Error())
	}

	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
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

	for _, value := range conf.Observer {
		if err := filepath.Walk(value.Path, watchDir); err != nil {
			fmt.Println("ERROR", err)
		}
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
				log.Println("event:", event)
				for _, cfg := range conf.Observer {
					if !strings.Contains(event.Name, string(cfg.Path)) {
						break
					}
					if !regexp.MustCompile(strings.Join(cfg.IncludeRegexp, "|")).MatchString(event.Name) {
						break
					}
					if regexp.MustCompile(strings.Join(cfg.ExcludeRegexp, "|")).MatchString(event.Name) {
						break
					}
					repos.RecordLog(event)
					log.Println("event:", event)
					if event.Has(fsnotify.Write) {
						log.Println("modified file:", event.Name)
					}

					for index, command := range cfg.Commands {
						temp = strings.Split(command, " ")
						cmd := exec.Command(temp[0], temp[1:]...)
						cmd.Dir = cfg.Path
						logError, err := cmd.CombinedOutput()
						if len(cfg.LogFile) != 0 && len(string(logError)) != 0 {
							writeLogToFile(cfg.Path+cfg.LogFile, string(logError))
						}
						if err != nil {
							log.Printf("Command №%d in path %s failed, skipping the rest \n", index+1, cfg.Path)
							break
						}
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

func writeLogToFile(path string, logError any) {

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	logInfo := log.New(f, "INFO\t", log.LstdFlags)
	logInfo.Println(logError)
	f.Close()
}
