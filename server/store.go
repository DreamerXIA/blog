package main

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

// Profile 是主页展示的个人信息，单行数据。
type Profile struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	TechDirection string `json:"tech_direction"`
	LearningGoals string `json:"learning_goals"`
}

// Log 是一条成长记录，type 区分 work / study / daily / summary。
type Log struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

var ErrNotFound = errors.New("not found")

var validLogTypes = map[string]bool{
	"work":    true,
	"study":   true,
	"daily":   true,
	"summary": true,
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.seed(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			type       TEXT NOT NULL,
			title      TEXT NOT NULL,
			content    TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS profile (
			id             INTEGER PRIMARY KEY CHECK (id = 1),
			name           TEXT NOT NULL,
			phone          TEXT NOT NULL,
			email          TEXT NOT NULL,
			tech_direction TEXT NOT NULL,
			learning_goals TEXT NOT NULL
		);
	`)
	return err
}

func (s *Store) seed() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM profile`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := s.db.Exec(`
		INSERT INTO profile (id, name, phone, email, tech_direction, learning_goals)
		VALUES (1, '你的名字', '', '', '', '');
	`)
	return err
}

func (s *Store) GetProfile() (Profile, error) {
	var p Profile
	err := s.db.QueryRow(`
		SELECT name, phone, email, tech_direction, learning_goals
		FROM profile WHERE id = 1
	`).Scan(&p.Name, &p.Phone, &p.Email, &p.TechDirection, &p.LearningGoals)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *Store) UpdateProfile(p Profile) error {
	_, err := s.db.Exec(`
		UPDATE profile
		SET name = ?, phone = ?, email = ?, tech_direction = ?, learning_goals = ?
		WHERE id = 1
	`, p.Name, p.Phone, p.Email, p.TechDirection, p.LearningGoals)
	return err
}

func (s *Store) CreateLog(logType, title, content string) (Log, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
		INSERT INTO logs (type, title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, logType, title, content, now, now)
	if err != nil {
		return Log{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Log{}, err
	}
	return Log{ID: id, Type: logType, Title: title, Content: content, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Store) ListLogs(logType string) ([]Log, error) {
	query := `SELECT id, type, title, content, created_at, updated_at FROM logs`
	args := []any{}
	if logType != "" {
		query += ` WHERE type = ?`
		args = append(args, logType)
	}
	query += ` ORDER BY id DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []Log{}
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.Type, &l.Title, &l.Content, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *Store) GetLog(id int64) (Log, error) {
	var l Log
	err := s.db.QueryRow(`
		SELECT id, type, title, content, created_at, updated_at
		FROM logs WHERE id = ?
	`, id).Scan(&l.ID, &l.Type, &l.Title, &l.Content, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return l, ErrNotFound
	}
	return l, err
}
