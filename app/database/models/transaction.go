package models

import "time"

type BackupTransaction struct {
	File      string    `json:"file"`
	Directory *string   `json:"directory"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
