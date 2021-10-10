package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"mysql-backup/app/USBChecker"
	"mysql-backup/app/database"
	"mysql-backup/config"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/xerrors"
)

type BackupHandler struct {
	db  *database.Database
	env *config.Env
}

func NewBackupHandler(db *database.Database, env *config.Env) *BackupHandler {
	return &BackupHandler{
		db:  db,
		env: env,
	}
}

func (h *BackupHandler) Backup(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "filename")
	err := h.db.InsertBackupTransaction(fileName)
	if err != nil {
		log.Printf("InsertBackupTransaction err: %+v", err)
		http.Error(w, "INTERNAL SERVER ERROR", http.StatusInternalServerError)
		return
	}

	dirPath, err := USBChecker.BackupDirMount(h.env.BackupDir)
	if err != nil {
		log.Printf("checkBackupDirMount err: %+v", err)
		http.Error(w, "INTERNAL SERVER ERROR", http.StatusInternalServerError)

		var mErr *USBChecker.MountError

		if xerrors.As(err, &mErr) {
			// transactionステータス更新(restarting)
			if err := h.db.UpdateStatusToRestart(fileName); err != nil {
				log.Printf("Error updating backup_transaction status: %+v", err)
			}
			// USBが接続されているが、正常にマウントされていないためpodを再起動してマウントし直す
			ProgramRestart("USB is connected, but not mounted, so restart and remount")
			return
		}
		if err := h.db.UpdateStatusToFailed(fileName); err != nil {
			log.Printf("Error updating backup_transaction status: %+v", err)
		}
		return
	}

	// ToDo: transactionステータス更新(success)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		{
			"message": "backup completed"
		}
	`))

	log.Printf("MountPoint is %s", dirPath)
	log.Printf("FileName is %s", fileName)

	if err := h.db.UpdateDirPath(fileName, dirPath); err != nil {
		log.Printf("Error updating backup_transaction dir_path: %+v", err)
		return
	}

	err = h.db.DumpDatabase(fileName, dirPath)
	if err != nil {
		log.Printf("backup error: %v", err)
		// transactionステータス更新(failed)
		if err := h.db.UpdateStatusToFailed(fileName); err != nil {
			log.Printf("Error updating backup_transaction status: %+v", err)
		}
		http.Error(w, "INTERNAL SERVER ERROR", 500)
		return
	}

	if err := h.db.UpdateStatusToSuccess(fileName); err != nil {
		log.Printf("Error updating backup_transaction status: %+v", err)
		return
	}
	log.Printf("finish backup correctry")
}

func (h *BackupHandler) GetDirectory(w http.ResponseWriter, r *http.Request) {
	directory, err := USBChecker.BackupDirMount(h.env.BackupDir)
	if err != nil {
		log.Printf("checkBackupDirMount err: %+v", err)
		var ncErr *USBChecker.NotConnect
		var mErr *USBChecker.MountError

		switch {
		case xerrors.As(err, &ncErr):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				{
					"directory": null,
					"mounted": false,
					"connected": false
				}
			`))
			return
		case xerrors.As(err, &mErr):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				{
					"directory": null,
					"mounted": false,
					"connected": true
				}
			`))
			// USBが接続されているが、正常にマウントされていないためpodを再起動してマウントし直す
			ProgramRestart("USB is connected, but not mounted, so restart and remount")
			return
		default:
			http.Error(w, "INTERNAL SERVER ERROR", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
		{
			"directory": "%v",
			"mounted": true,
			"connected": true
		}
	`, filepath.Base(directory))))
}

func (h *BackupHandler) Status(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "filename")
	info, err := h.db.GetTransactionInfo(fileName)
	if err != nil {
		log.Printf("cannot get transaction info: %+v", err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
		return
	}
	*info.Directory = filepath.Base(*info.Directory)

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("structure parse error: %+v", err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (h *BackupHandler) Restart(w http.ResponseWriter, r *http.Request) {
	ProgramRestart()
}
