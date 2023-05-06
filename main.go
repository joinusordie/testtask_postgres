package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
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


    // Create new watcher.
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

    if err := filepath.Walk(viper.GetString("path"), watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

    done := make(chan bool)

    // Start listening for events.
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                log.Println("event:", event)
                if event.Has(fsnotify.Write) {
                    log.Println("modified file:", event.Name)
                }
                for _, value:= range viper.GetStringSlice("commands"){
                    temp = strings.Split(value, " ")
                    cmd := exec.Command(temp[0], temp[1:]...)
                    cmd.Dir = viper.GetString("path")
                    fmt.Println(cmd)
                    err := cmd.Run()
                    if err != nil {
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


    <-done
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