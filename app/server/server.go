package server

import (
	"net/http"

	"mysql-backup/app/database"
	"mysql-backup/app/server/handlers"
	"mysql-backup/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/xerrors"
)

type Server struct {
	db     *database.Database
	env    *config.Env
	router *chi.Mux
}

func NewServer(db *database.Database, env *config.Env) *Server {
	return &Server{
		db:     db,
		env:    env,
		router: chi.NewRouter(),
	}
}

func (s *Server) Run() error {
	// TODO サーバーが立ち上がる前に、トランザクションテーブルをみて、前回中断したと判断したら、
	// backup を実行したほうがよい。

	h := handlers.NewBackupHandler(s.db, s.env)
	s.setRouter(h)

	err := http.ListenAndServe(":30088", s.router)
	if err != nil {
		return xerrors.Errorf("An error occurred during runnning server: %w", err)
	}
	return nil
}

func (s *Server) setRouter(h *handlers.BackupHandler) {
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	s.router.Route("/api/backup", func(r chi.Router) {
		r.Post("/{filename}", h.Backup)
		r.Get("/status/{filename}", h.Status)
		r.Get("/directory", h.GetDirectory)
		r.Get("/restart", h.Restart)
	})
}

// Todo: /backupで実行される関数の内容を確認してもらう
// Todo: 旧backupでenvを読み込んでいる箇所をどうするか?
// 		 .envから読む or それ以外
