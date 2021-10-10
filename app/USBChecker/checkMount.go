package USBChecker

import (
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

// BackupDirMount 接続されているUSBのマウントポイントを返す
func BackupDirMount(backupDir string) (string, error) {
	dirs, err := dirsOnBackupDir(backupDir)
	if err != nil {
		return "", xerrors.Errorf("BackupDir err: %w", err)
	}

	dir, err := findDirMountedOnSDx1(dirs)
	if err != nil {
		return "", xerrors.Errorf("FindDirMountedOnSDA1 err: %w", err)
	}

	return dir, nil
}

// dirsOnBackupDir backup directory に存在する directory のフルパス一覧を取得する。
func dirsOnBackupDir(backupDir string) ([]string, error) {
	items, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return nil, xerrors.Errorf("failed to get dir info: %w", err)
	}
	dirs := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			dirs = append(dirs, filepath.Join(backupDir, item.Name()))
		}
	}
	return dirs, nil
}

// findDirMountedOnSDx1 受け取った directory フルパスの内、SDA1がマウントしているパスを確認
func findDirMountedOnSDx1(dirpaths []string) (string, error) {
	cmd := "lsblk | grep -E '─sd.1 ' | awk '{print $1,$7}'"
	log.Printf("command is %v\n", cmd)

	// exec.Commandは以下のように指定しないと動作しない
	_cmdRes, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", xerrors.Errorf("failed to execute command: %w", err)
	}
	log.Printf("output is %v\n", string(_cmdRes))

	spaceRE := regexp.MustCompile(`^\s|^()$`)
	if spaceRE.Match(_cmdRes) {
		return "", &NotConnect{message: "SDA1 does not exist"}
	}

	cmdRes := strings.Split(string(_cmdRes), " ")
	if len(cmdRes) < 2 {
		return "", xerrors.Errorf("command result is unexpected : %v", cmdRes)
	}

	mountPath := strings.Trim(cmdRes[1], "\n")
	if mountPath == "" {
		return "", &MountError{message: "SDA1 is not mounted"}
	}

	for _, dir := range dirpaths {
		if dir == mountPath {
			return dir, nil
		}
	}

	return "", xerrors.Errorf("sda1 is mounted on %s, but specified dirs has not the dir name %v", mountPath, dirpaths)
}
