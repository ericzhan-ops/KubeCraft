package main

import (
	"KubeCraft/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"KubeCraft/internal/deploy"
	"KubeCraft/internal/initialize"
)

// corsMiddleware 添加CORS支持的中间件
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头部
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 如果是OPTIONS预检请求，直接返回
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 否则继续处理请求
		next(w, r)
	}
}

// SSEProgressReporter SSE进度报告器
type SSEProgressReporter struct {
	writer http.ResponseWriter
	step   int
	total  int
}

// ReportProgress 报告进度
func (s *SSEProgressReporter) ReportProgress(message string) {
	s.step++
	progress := ProgressMessage{
		Type:    "progress",
		Message: message,
		Step:    s.step,
		Total:   s.total,
	}

	data, _ := json.Marshal(progress)
	_, err := fmt.Fprintf(s.writer, "data: %s\n\n", string(data))
	if err != nil {
		return
	}
	s.writer.(http.Flusher).Flush()

	// 短暂延迟以模拟处理时间
	time.Sleep(500 * time.Millisecond)
}

func main() {
	// 注册API路由处理器，添加CORS支持
	http.HandleFunc("/api/init/progress", corsMiddleware(initializeProgress))
	http.HandleFunc("/api/deploy/progress", corsMiddleware(deployProgress))

	// 提供静态文件服务
	http.HandleFunc("/", serveStaticFiles)

	// 启动服务器
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// serveStaticFiles 处理静态文件服务
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	// 如果请求的是API路径，则返回404
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// 默认访问根路径时返回index.html
	if r.URL.Path == "/" {
		// 为首页也添加CORS支持
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "index.html")
		return
	}

	// 其他静态文件
	http.ServeFile(w, r, r.URL.Path[1:])
}

// ProgressMessage 进度消息结构
type ProgressMessage struct {
	Type    string `json:"type"`    // "progress" 或 "complete" 或 "error"
	Message string `json:"message"` // 消息内容
	Step    int    `json:"step"`    // 当前步骤
	Total   int    `json:"total"`   // 总步骤数
}

// initializeProgress 处理初始化进度推送
func initializeProgress(w http.ResponseWriter, r *http.Request) {
	// 设置SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 解析请求体中的配置
	var config utils.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		// 如果没有请求体，使用默认配置
		config = utils.Config{}
	}

	// 创建进度报告器
	reporter := &SSEProgressReporter{
		writer: w,
		step:   0,
		total:  len(initialize.Playbooks) + 3, // 3个额外步骤：检查ansible、安装ansible、复制配置
	}

	// 执行初始化过程
	err := initialize.Process(config, reporter)

	// 发送完成或错误消息
	if err != nil {
		errorMsg := ProgressMessage{
			Type:    "error",
			Message: fmt.Sprintf("初始化失败: %v", err),
			Step:    reporter.step,
			Total:   reporter.total,
		}
		data, _ := json.Marshal(errorMsg)
		_, err := fmt.Fprintf(w, "data: %s\n\n", string(data))
		if err != nil {
			return
		}
		w.(http.Flusher).Flush()
	} else {
		complete := ProgressMessage{
			Type:    "complete",
			Message: "集群初始化完成",
			Step:    reporter.step,
			Total:   reporter.total,
		}
		data, _ := json.Marshal(complete)
		_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
		if err != nil {
			return
		}
		w.(http.Flusher).Flush()
	}
}

// deployProgress 处理部署进度推送
func deployProgress(w http.ResponseWriter, r *http.Request) {
	// 设置SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 解析请求体中的配置
	var config utils.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		// 如果没有请求体，使用默认配置
		config = utils.Config{}
	}

	// 创建进度报告器
	reporter := &SSEProgressReporter{
		writer: w,
		step:   0,
		total:  len(deploy.Playbooks) + 1, // 1个额外步骤：安装附加组件
	}

	// 执行部署过程
	err := deploy.Process(config, reporter)

	// 发送完成或错误消息
	if err != nil {
		errorMsg := ProgressMessage{
			Type:    "error",
			Message: fmt.Sprintf("部署失败: %v", err),
			Step:    reporter.step,
			Total:   reporter.total,
		}
		data, _ := json.Marshal(errorMsg)
		_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
		if err != nil {
			return
		}
		w.(http.Flusher).Flush()
	} else {
		complete := ProgressMessage{
			Type:    "complete",
			Message: "集群部署完成",
			Step:    reporter.step,
			Total:   reporter.total,
		}
		data, _ := json.Marshal(complete)
		_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
		if err != nil {
			return
		}
		w.(http.Flusher).Flush()
	}
}
