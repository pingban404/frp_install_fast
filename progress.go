package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// ProgressReader 带进度显示的Reader
type ProgressReader struct {
	io.Reader
	total    int64
	current  int64
	progress chan ProgressUpdate
}

// ProgressUpdate 进度更新信息
type ProgressUpdate struct {
	Current int64
	Total   int64
	Percent float64
}

// NewProgressReader 创建新的进度Reader
func NewProgressReader(r io.Reader, total int64) *ProgressReader {
	return &ProgressReader{
		Reader:   r,
		total:    total,
		progress: make(chan ProgressUpdate, 1),
	}
}

// Read 实现io.Reader接口，同时更新进度
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.current, int64(n))
		current := atomic.LoadInt64(&pr.current)
		percent := float64(current) / float64(pr.total) * 100
		
		// 非阻塞发送进度更新
		select {
		case pr.progress <- ProgressUpdate{Current: current, Total: pr.total, Percent: percent}:
		default:
		}
	}
	return n, err
}

// StartProgress 启动进度显示
func (pr *ProgressReader) StartProgress(description string, fm *FrpsManager) {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // 每100ms更新一次
		defer ticker.Stop()
		
		var lastUpdate ProgressUpdate
		for {
			select {
			case update := <-pr.progress:
				lastUpdate = update
			case <-ticker.C:
				if lastUpdate.Total > 0 {
					fm.showProgress(description, lastUpdate)
				}
			}
			
			// 下载完成时退出
			if lastUpdate.Current >= lastUpdate.Total && lastUpdate.Total > 0 {
				fm.showProgress(description, lastUpdate)
				fmt.Println() // 换行
				return
			}
		}
	}()
}

// showProgress 显示进度条
func (fm *FrpsManager) showProgress(description string, update ProgressUpdate) {
	if update.Total == 0 {
		return
	}
	
	barWidth := 40
	filled := int(update.Percent / 100 * float64(barWidth))
	
	// 构建进度条
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	
	// 格式化文件大小
	current := fm.formatBytes(update.Current)
	total := fm.formatBytes(update.Total)
	
	// 显示进度 (使用\r回到行首覆盖之前的内容)
	fmt.Printf("\r%s [%s] %.1f%% (%s/%s)", 
		description, bar, update.Percent, current, total)
}

// formatBytes 格式化字节数为人类可读格式
func (fm *FrpsManager) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// downloadWithProgress 带进度条的下载函数
func (fm *FrpsManager) downloadWithProgress(url, filename, description string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// 获取文件大小
	total := resp.ContentLength
	if total <= 0 {
		total = 1 // 避免除零错误，对于未知大小的文件
	}
	
	// 创建目标文件
	tmpFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	
	// 创建进度Reader
	progressReader := NewProgressReader(resp.Body, total)
	progressReader.StartProgress(description, fm)
	
	// 复制数据
	_, err = io.Copy(tmpFile, progressReader)
	if err != nil {
		return err
	}
	
	// 等待进度显示完成
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

// downloadWithProgressForScript 为脚本下载提供的带进度条下载功能
func (fm *FrpsManager) downloadWithProgressForScript(url, filename, description string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// 获取文件大小
	total := resp.ContentLength
	if total <= 0 {
		// 对于脚本文件，如果没有Content-Length，假设一个小的默认值
		total = 8192 // 8KB
	}
	
	// 创建目标文件
	tmpFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	
	// 由于脚本文件通常很小，我们使用一个简化的进度显示
	if total > 1024*10 { // 只有大于10KB的文件才显示进度条
		// 创建进度Reader
		progressReader := NewProgressReader(resp.Body, total)
		progressReader.StartProgress(description, fm)
		
		// 复制数据
		_, err = io.Copy(tmpFile, progressReader)
		if err != nil {
			return err
		}
		
		// 等待进度显示完成
		time.Sleep(200 * time.Millisecond)
	} else {
		// 对于小文件，直接复制不显示进度条
		fmt.Printf("%s...", description)
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(" 完成")
	}
	
	return nil
} 