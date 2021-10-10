package handlers

import (
	"log"
	"os"
)

// ProgramRestart kube上で動かす想定だから、
// プログラムを終了させれば勝手に再立ち上げされる。
func ProgramRestart(msg ...interface{}) {

	// TODO なにか良い方法があれば、変更してください。
	log.Print(msg...)
	os.Exit(0)
}
