package config

import (
	"fmt"
	"os"
)

const USBPATH = "/media/latona/"

type Env struct {
	*MysqlEnv
	BackupDir string
}

func NewEnv() *Env {
	return &Env{
		MysqlEnv:  NewMysqlEnv(),
		BackupDir: getEnv("BACKUP_DIR", "/media/latona/"),
	}
}

type MysqlEnv struct {
	User     string
	Host     string
	Password string
	Port     string
}

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		value = def
	}
	return value
}

func NewMysqlEnv() *MysqlEnv {
	user := getEnv("MYSQL_USER", "XXXXXX")
	host := getEnv("MYSQL_HOST", "mysql")
	pass := getEnv("MYSQL_PASSWORD", "XXXXXXXXX")
	port := getEnv("MYSQL_PORT", "XXXX")

	return &MysqlEnv{
		User:     user,
		Host:     host,
		Password: pass,
		Port:     port,
	}
}

func (c *MysqlEnv) DSN() string {
	return fmt.Sprintf(`XXXXXXXXXXXXXXXxx`, c.User, c.Password, c.Host, c.Port, "XXXXXXXX")
}
