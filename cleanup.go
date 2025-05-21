// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vortexrotate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

// CleanUp 根据文件最大数量来确定是否执行清理
type CleanUp struct {
	// 文件所在目录
	dir string
	// 最大数量
	maxCount uint64
	// 保存的周期
	period uint16
	// timer
	timer *time.Timer
	// ticker
	ticker *time.Ticker
	// 关闭信号
	sig chan struct{}
	// 检查的时间间隔
	interval time.Duration
	// 正则匹配
	re *regexp.Regexp
	// 加锁保护
	lock sync.RWMutex
}

func NewFileCountCleanUp(dir, filename string, maxCount uint64, period uint16) *CleanUp {
	// 正则匹配文件名中的日期和序号
	escapedPrefix := regexp.QuoteMeta(filename)
	fileNameRegexPattern := fmt.Sprintf(`%s_(\d{8})_(\d{4})\.log`, escapedPrefix)
	fc := CleanUp{
		dir:      dir,
		maxCount: maxCount,
		period:   period,
		sig:      make(chan struct{}),
		lock:     sync.RWMutex{},
		re:       regexp.MustCompile(fileNameRegexPattern),
	}

	return &fc
}

func (c *CleanUp) Start() {
	c.lock.Lock()
	defer c.lock.Unlock()
	const TimeDuration = 24 * time.Hour
	c.timer = time.AfterFunc(TimeDuration, func() {
		c.cleanExpiredFiles()
		go c.startTicker()
	})
}

func (c *CleanUp) startTicker() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.sig:
			return
		case <-ticker.C:
			c.cleanExpiredFiles()
		}
	}
}

func (c *CleanUp) ResetInterval(newInterval time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.interval = newInterval
	if c.ticker != nil {
		c.ticker.Reset(newInterval)
	}
}

func (c *CleanUp) cleanExpiredFiles() {
	list, err := c.listFileInfo()
	if err != nil {
		// TODO 处理错误
		return
	}
	if len(list) == 0 {
		return
	}
	fileInfos := make([]string, 0, len(list))
	for _, info := range list {
		fileInfos = append(fileInfos, info.Name())
	}
	_, err = c.sortFiles(fileInfos)
	if err != nil {
		// TODO 处理错误
		return
	}

	// 执行删除
}

func (c *CleanUp) listFileInfo() ([]os.FileInfo, error) {
	var logFiles []os.FileInfo
	err := filepath.WalkDir(c.dir, func(_ string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		if fileInfo != nil {
			logFiles = append(logFiles, fileInfo)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return logFiles, nil
}

func (c *CleanUp) sortFiles(files []string) ([]FileInfo, error) {
	fileInfos := make([]FileInfo, 0, len(files))
	const matchesLen = 3
	for _, f := range files {
		matches := c.re.FindStringSubmatch(f)
		if len(matches) < matchesLen {
			return fileInfos, fmt.Errorf("regrexp find sub match error, filename: %s", f)
		}

		date := matches[1]
		t, err := time.Parse(Layout, date)
		if err != nil {
			return fileInfos, fmt.Errorf("regrexp parse date error, filename: %s, date: %s", f, matches[1])
		}

		sequence, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return fileInfos, fmt.Errorf("parse sequence error, filename: %s, sequence: %s", f, matches[2])
		}
		fileInfos = append(fileInfos, FileInfo{
			UpDir:    fmt.Sprintf("%s/%s", c.dir, date),
			Name:     f,
			Date:     t,
			Sequence: sequence,
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		if !fileInfos[i].Date.Equal(fileInfos[j].Date) {
			// 不同日期的文件
			return fileInfos[i].Date.Before(fileInfos[j].Date)
		}

		// 相同日志的文件比对序列号
		return fileInfos[i].Sequence < fileInfos[j].Sequence
	})

	return fileInfos, nil
}

func (c *CleanUp) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()
	close(c.sig)
	if c.timer != nil {
		c.timer.Stop()
	}
}

// FileInfo 文件信息
type FileInfo struct {
	UpDir    string    // 父目录
	Name     string    // 文件名称
	Date     time.Time // 文件时间(年月日)
	Sequence int64     // 文件序列号
}
