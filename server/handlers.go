package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type server struct {
	store *Store
	token string
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/profile", s.handleGetProfile)
	mux.HandleFunc("PUT /api/profile", s.handleUpdateProfile)
	mux.HandleFunc("GET /api/logs", s.handleListLogs)
	mux.HandleFunc("POST /api/logs", s.handleCreateLog)
	mux.HandleFunc("GET /api/logs/{id}", s.handleGetLog)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProfile()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取个人信息失败")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		writeError(w, http.StatusUnauthorized, "缺少或错误的 token")
		return
	}
	var p Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		writeError(w, http.StatusBadRequest, "姓名不能为空")
		return
	}
	p.Phone = strings.TrimSpace(p.Phone)
	p.Email = strings.TrimSpace(p.Email)
	p.TechDirection = strings.TrimSpace(p.TechDirection)
	p.LearningGoals = strings.TrimSpace(p.LearningGoals)

	if err := s.store.UpdateProfile(p); err != nil {
		writeError(w, http.StatusInternalServerError, "保存个人信息失败")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *server) handleListLogs(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("type")
	if logType != "" && !validLogTypes[logType] {
		writeError(w, http.StatusBadRequest, "无效的日志类型")
		return
	}
	logs, err := s.store.ListLogs(logType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取日志失败")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *server) handleCreateLog(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		writeError(w, http.StatusUnauthorized, "缺少或错误的 token")
		return
	}
	var body struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}
	body.Type = strings.TrimSpace(body.Type)
	body.Title = strings.TrimSpace(body.Title)
	body.Content = strings.TrimSpace(body.Content)

	if !validLogTypes[body.Type] {
		writeError(w, http.StatusBadRequest, "日志类型必须是 work/study/daily/summary")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "标题不能为空")
		return
	}
	if body.Content == "" {
		writeError(w, http.StatusBadRequest, "正文不能为空")
		return
	}

	log, err := s.store.CreateLog(body.Type, body.Title, body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "保存日志失败")
		return
	}
	writeJSON(w, http.StatusCreated, log)
}

func (s *server) handleGetLog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的日志 id")
		return
	}
	log, err := s.store.GetLog(id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "日志不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取日志失败")
		return
	}
	writeJSON(w, http.StatusOK, log)
}

func (s *server) authorized(r *http.Request) bool {
	if s.token == "" {
		return false
	}
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	return strings.TrimPrefix(auth, prefix) == s.token
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
