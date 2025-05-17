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
	"github.com/TimeWtr/vortexrotate/errorx"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"sync"
	"time"
)

const (
	RotateSizeThreshold = 0.8
	RotateInterval      = time.Millisecond * 100
)

type TimingType string

const (
	_Second TimingType = "second"
	Hour    TimingType = "hour"
	Day     TimingType = "day"
	Week    TimingType = "week"
	Month   TimingType = "month"
)

func (t TimingType) String() string {
	return string(t)
}

func (t TimingType) Valid() bool {
	switch t {
	case _Second, Hour, Day, Week, Month:
		return true
	default:
		return false
	}
}

// RotateStrategy 文件轮转的策略
type RotateStrategy interface {
	// ShouldRotate 写入文件时立即判断否应该执行文件轮转
	ShouldRotate(writeSize uint64) bool
	// NotifyRotate 获取定时轮转信号
	NotifyRotate() <-chan struct{}
	// Close 关闭轮转策略
	Close()
}

var _ RotateStrategy = (*MixStrategy)(nil)

// MixStrategy 混合策略包括两个触发因子：定时和当前文件大小。
// 文件大小的优先级高于定时，当在一个定时的时间窗口周期内，数据写入时文件
// 的大小超过了最大限制，比如：100M，则立即返回需要执行轮转操作，由调用方
// 立即执行轮转。当定时任务达到轮转时间后，发送一条轮转事件通知到通知通道，
// 调用方实时监听信号，收到信号信号后执行轮转操作。每次轮转都需要记录当前轮
// 转的时间戳，如果是文件大小超过限制而触发的轮转，则无须进行时间判断，立即
// 进行轮转，如果是定时任务触发的轮转，则需要比较当前时间的时间和上次轮转的
// 时间的差值是否超过一定的时间范围，比如10分钟，或当前文件的写入大小是否达
// 到了最大大小的一定比例，比如80%，如果满足其中一条才能进行轮转操作，发送
// 事件通知，反之则跳过本次定时轮转。
type MixStrategy struct {
	// 单个文件允许的最大字节
	maxSize uint64
	// 当前已经写入的最大字节数
	size uint64
	// 加锁保护
	lock sync.Mutex
	// 定时任务
	c *cron.Cron
	// 定时事件类型
	tp TimingType
	// 上次轮转的事件
	lastTime int64
	// 事件通知通道，只用于定时轮转的事件通知
	events chan struct{}
	// 日志
	lg *log.Logger
}

func NewMixStrategy(maxSize uint64, tp TimingType) (*MixStrategy, error) {
	if !tp.Valid() {
		return nil, errorx.ErrTimeType
	}

	stg := &MixStrategy{
		maxSize: maxSize,
		lock:    sync.Mutex{},
		events:  make(chan struct{}),
		c:       cron.New(cron.WithSeconds()),
		tp:      tp,
		lg:      log.New(os.Stdout, "", log.LstdFlags),
	}

	if err := stg.asyncWorker(); err != nil {
		return nil, err
	}

	return stg, nil
}

func (s *MixStrategy) NotifyRotate() <-chan struct{} {
	return s.events
}

// ShouldRotate 是否需要执行轮转操作
func (s *MixStrategy) ShouldRotate(writeSize uint64) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.size+writeSize < s.maxSize {
		s.size += writeSize
		return false
	}

	s.lastTime = time.Now().UnixMilli()
	s.size = 0
	return true
}

// asyncWorker 开启定时任务执行轮转判断逻辑，定时任务表达式根据定时任务类型确定
// _Second: 不支持秒级的定时任务，这个只用于单元测试
// Hour: 每隔一小时执行一次，0 0 * * * *
// Day: 每天凌晨0点执行一次，0 0 0 * * *
// Week: 每周一凌晨0点执行一次，0 0 0 * * 1
// Month: 每月1号凌晨0点执行一次，0 0 0 1 * *
func (s *MixStrategy) asyncWorker() error {
	var cronStr string
	switch s.tp {
	case _Second:
		cronStr = "*/1 * * * * *"
	case Hour:
		cronStr = "0 0 * * * *"
	case Day:
		cronStr = "0 0 0 * * *"
	case Week:
		cronStr = "0 0 0 * * 1"
	case Month:
		cronStr = "0 0 0 1 * *"
	default:
		return errorx.ErrTimeType
	}

	_, err := s.c.AddFunc(cronStr, func() {
		s.lock.Lock()
		if time.Duration(time.Now().UnixMilli()-s.lastTime) < RotateInterval {
			if float64(s.size) < float64(s.maxSize)*RotateSizeThreshold {
				threshold := float64(s.maxSize) * RotateSizeThreshold
				s.lg.Printf("rotate size too small, size: %d, threshold: %0.2f, skip!", s.size, threshold)
				s.lock.Unlock()
				return
			}
		}

		s.lastTime = time.Now().UnixMilli()
		s.size = 0
		s.lock.Unlock()

		select {
		case s.events <- struct{}{}:
			s.lg.Println("rotate event send success!")
		case <-time.After(time.Second):
			s.lg.Println("rotate event send timeout!")
		}
	})

	s.c.Start()

	return err
}

func (s *MixStrategy) Close() {
	s.c.Stop()
	close(s.events)
}
