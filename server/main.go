package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// resolveStaticDir 返回首个存在的前端静态目录。按优先级：显式 STATIC_DIR，
// 然后是相对当前工作目录的常见位置（从 server/ 或仓库根启动均可命中）。
func resolveStaticDir() string {
	if v := os.Getenv("STATIC_DIR"); v != "" {
		return v
	}
	for _, candidate := range []string{
		filepath.Join("..", "web", "dist"),
		filepath.Join("web", "dist"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func main() {
	port := envOr("PORT", "8080")
	token := envOr("OWNER_TOKEN", "dev-token")
	dbPath := envOr("DB_PATH", "blog.db")

	store, err := Open(dbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer store.Close()

	s := &server{store: store, token: token}
	mux := s.routes()

	if staticDir := resolveStaticDir(); staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fs)
		log.Printf("托管前端静态文件: %s", staticDir)
	} else {
		log.Printf("未找到前端静态目录，仅提供 API（开发模式下请用 Vite dev server）")
	}

	addr := ":" + port
	log.Printf("服务启动于 http://localhost%s（token 保护写入）", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
