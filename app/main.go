package main

import (
	"log"

	"mysql-backup/app/USBChecker"
	"mysql-backup/app/database"
	"mysql-backup/app/server"
	"mysql-backup/config"
)

func main() {
	log.Printf("server start")
	defer log.Printf("server finish")

	env := config.NewEnv()
	db := database.NewDatabase(env.MysqlEnv)
	err := db.Connect()
	if err != nil {
		log.Printf("failed to create database: %v", err)
		return
	}
	defer db.Disconnect()

	dirPath, err := USBChecker.BackupDirMount(env.BackupDir)
	if err != nil {
		log.Printf("can not get mount point. BUT CONTINUE: %+v", err)
	} else {
		err = db.RetryDumpDatabase(dirPath)
		if err != nil {
			log.Printf("can not Retry dump database. BUT CONTINUE: %+v", err)
		}
	}

	s := server.NewServer(db, env)
	err = s.Run()
	if err != nil {
		log.Printf("Run server err: %v", err)
		return
	}
}
