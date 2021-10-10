package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"mysql-backup/app/database/models"

	"github.com/JamesStewy/go-mysqldump"
	"golang.org/x/xerrors"
)

type Status int

const (
	STATUS_PROCESSING Status = iota + 1
	STATUS_FAILED
	STATUS_SUCCESS
	STATUS_RESTARTING
	// STATUS_NONE
)

func (s Status) String() string {
	switch s {
	case STATUS_PROCESSING:
		return "processing"
	case STATUS_FAILED:
		return "failed"
	case STATUS_SUCCESS:
		return "success"
	case STATUS_RESTARTING:
		return "restarting"
	default:
		return "unknown"
	}
}

func (d *Database) dumpDatabase(fileName, dirPath string) error {
	preFileName := fmt.Sprintf("XXXXXXXXXXX", "before-rename")
	dumper, err := mysqldump.Register(d.db, dirPath, preFileName)
	if err != nil {
		return xerrors.Errorf("Error registering databse: %w", err)
	}
	// Dump database to file
	resultFileName, err := dumper.Dump()
	if err != nil {
		return xerrors.Errorf("Error dumping: %w", err)
	}

	err = changeFileName(dirPath, filepath.Base(resultFileName), fileName+".sql")
	if err != nil {
		log.Printf("can not change filename : %+v", err)
		log.Printf("File is saved to %s", resultFileName)
		return nil
	}

	log.Printf("File is saved to %s", resultFileName)
	return nil
}

func (d *Database) retryDumpDatabase(dir string) error {
	restartRows, err := d.getBackupTransactionByStatus(STATUS_RESTARTING)
	if err != nil {
		return xerrors.Errorf("cannot get rows status %s: %w", STATUS_RESTARTING, err)
	}

	if len(restartRows) == 0 {
		return nil
	}

	PKs := make([]string, 0, len(restartRows))
	for _, r := range restartRows {
		PKs = append(PKs, r.File)
	}

	if err := d.bulkUpdateStatusToFailed(PKs); err != nil {
		log.Printf("[DANGER] %+v", xerrors.Errorf("cannot update %d rows' status to FAILED. its PK is %v : %w", len(PKs), PKs, err))
	} else {
		log.Printf("update %d rows' status to FAILED", len(PKs))
	}

	restartRow := restartRows[0]

	err = d.DumpDatabase(dir, restartRow.File)
	if err != nil {
		return xerrors.Errorf("dump failed: %w", err)
	}

	err = d.updateCustomStatus(restartRow.File, STATUS_SUCCESS)
	if err != nil {
		return xerrors.Errorf("cannot update status to SUCCESS: %w", err)
	}
	return nil
}

func (d *Database) bulkUpdateStatusToFailed(fileNames []string) error {
	if len(fileNames) == 0 {
		return nil
	}

	q := fmt.Sprintf(`
	UPDATE backup_transaction
	SET status = '%s'
	WHERE 1=1
	`, STATUS_FAILED)

	for _, fn := range fileNames {
		add := fmt.Sprintf("OR file_name = '%s'", fn)
		q = fmt.Sprintln(q, add)
	}

	q = fmt.Sprintln(q, ";")

	_, err := d.db.Exec(q)
	if err != nil {
		return xerrors.Errorf("update transaction status failed: %w", err)
	}
	return nil

}

func (d *Database) getBackupTransactionByStatus(status Status) ([]*models.BackupTransaction, error) {
	if err := d.db.Ping(); err != nil {
		return nil, xerrors.Errorf("DB ping err: %w", err)
	}

	rows, err := d.db.Query(
		fmt.Sprintf("SELECT * FROM backup_transaction WHERE status = '%s' ORDER BY timestamp desc", status),
	)
	if err != nil {
		return nil, xerrors.Errorf("query exec error: %w", err)
	}
	defer rows.Close()

	var t *models.BackupTransaction
	var result []*models.BackupTransaction
	for rows.Next() {
		err := rows.Scan(&t.File, &t.Directory, &t.Status, &t.Timestamp)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if len(result) == 0 {
		log.Printf("cannot find %s status rows, does not exist. BUT CONTINUE", status)
	}
	return result, nil
}

func (d *Database) getBackupTransactionByFileName(fileName string) (*models.BackupTransaction, error) {
	if err := d.db.Ping(); err != nil {
		return nil, xerrors.Errorf("DB ping err: %w", err)
	}

	rows, err := d.db.Query(
		fmt.Sprintf("SELECT * FROM backup_transaction WHERE file_name = %s ORDER BY timestamp desc", fileName),
	)
	if err != nil {
		return nil, xerrors.Errorf("query exec error: %w", err)
	}
	defer rows.Close()

	var t models.BackupTransaction
	var result []*models.BackupTransaction
	for rows.Next() {
		err := rows.Scan(&t.File, &t.Directory, &t.Status, &t.Timestamp)
		if err != nil {
			return nil, err
		}
		result = append(result, &t)
	}
	if len(result) == 0 {
		return nil, xerrors.Errorf("cannot find %s, does not exist", fileName)
	}
	return result[0], nil
}

func (d *Database) insertBackupTransaction(filename string) (time.Time, error) {
	ts := time.Now().UTC()
	_, err := d.db.Exec(
		fmt.Sprintf(`
		INSERT INTO backup_transaction
		(file_name, dir_path, status, timestamp)
		VALUE ('%s',null,'%s','%s');
		`, filename, STATUS_PROCESSING, ts.Format("2006-01-02 15:04:05")),
	)
	if err != nil {
		return time.Time{}, xerrors.Errorf("insert transaction failed: %w", err)
	}
	return ts, nil
}

func (d *Database) createRowToStartBackup(filename string) error {
	_, err := d.insertBackupTransaction(filename)
	if err != nil {
		return xerrors.Errorf("cannot update status to SUCCESS: %w", err)
	}
	return nil
}

func (d *Database) updateCustomStatus(fileName string, status Status) error {
	_, err := d.db.Exec(
		fmt.Sprintf(`
		UPDATE backup_transaction
		SET status = '%s'
		WHERE file_name = '%s';
		`, status, fileName),
	)
	if err != nil {
		return xerrors.Errorf("update transaction status failed: %w", err)
	}
	return nil
}

func (d *Database) updateDirPath(fileName string, dirPath string) error {
	_, err := d.db.Exec(
		fmt.Sprintf(`
		UPDATE backup_transaction
		SET dir_path = '%s'
		WHERE file_name = '%s';
		`, dirPath, fileName),
	)
	if err != nil {
		return xerrors.Errorf("update transaction dir_path failed: %w", err)
	}
	return nil
}

func changeFileName(dirName, fromFileName, toFileName string) error {
	src := filepath.Join(dirName, fromFileName)
	dst := filepath.Join(dirName, toFileName)

	err := os.Rename(src, dst)
	if err != nil {
		return xerrors.Errorf("failed to rename file : %w", err)
	}

	return nil
}
