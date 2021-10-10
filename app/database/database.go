package database

import (
	"database/sql"
	"log"
	"mysql-backup/app/database/models"
	"mysql-backup/config"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/xerrors"
)

type Database struct {
	db          *sql.DB
	config      *config.MysqlEnv
	transaction *models.BackupTransaction
}

func NewDatabase(mysqlEnv *config.MysqlEnv) *Database {
	return &Database{
		db:          nil,
		config:      mysqlEnv,
		transaction: nil,
	}
}

func (db *Database) Connect() error {
	d, err := sql.Open("mysql", db.config.DSN())
	if err != nil {
		return xerrors.Errorf(`failed to connect to database: %w`, err)
	}
	if err := d.Ping(); err != nil {
		return xerrors.Errorf("DB ping err: %w", err)
	}
	db.db = d
	return nil
}

func (d *Database) Disconnect() {
	err := d.db.Close()
	if err != nil {
		log.Printf("DB close err: %v", err)
	}
}

func (d *Database) DumpDatabase(fileName, dirPath string) error {
	return d.dumpDatabase(fileName, dirPath)
}

func (d *Database) RetryDumpDatabase(dir string) error {
	return d.retryDumpDatabase(dir)
}

func (d *Database) InsertBackupTransaction(filename string) error {
	_, err := d.insertBackupTransaction(filename)
	return err
}

func (d *Database) GetTransactionInfo(filename string) (*models.BackupTransaction, error) {
	return d.getBackupTransactionByFileName(filename)
}

func (d *Database) UpdateStatusToSuccess(filename string) error {
	err := d.updateCustomStatus(filename, STATUS_SUCCESS)
	if err != nil {
		return xerrors.Errorf("cannot update status to SUCCESS: %w", err)
	}
	return nil
}

func (d *Database) UpdateStatusToRestart(filename string) error {
	err := d.updateCustomStatus(filename, STATUS_RESTARTING)
	if err != nil {
		return xerrors.Errorf("cannot update status to RESTARTING: %w", err)
	}
	return nil
}

func (d *Database) UpdateStatusToFailed(filename string) error {
	err := d.updateCustomStatus(filename, STATUS_FAILED)
	if err != nil {
		return xerrors.Errorf("cannot update status to FAILED: %w", err)
	}
	return nil
}

func (d *Database) UpdateDirPath(filename string, dirPath string) error {
	err := d.updateDirPath(filename, dirPath)
	if err != nil {
		return xerrors.Errorf("cannot update dir_path: %w", err)
	}
	return nil
}
