package time

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"echat/utils/logger"
)

type SchedulerCallback func(interval time.Duration, deliverTime time.Time)

// Scheduler 定时任务
type Scheduler interface {
	// Start 启动
	Start(context context.Context, group *sync.WaitGroup) error
	// Stop 停止
	Stop()
	// Schedule 添加计时任务
	Schedule(duration time.Duration, isTicker bool, callback SchedulerCallback) (scheduleId uint64, err error)
	// Unschedule 停止计时任务
	Unschedule(scheduleId uint64) error
	// Done 定时回调 chan
	Done() <-chan *DeliverInfo
}

// DeliverInfo 任务回调信息
type DeliverInfo struct {
	scheduleId  uint64
	deliverTime time.Time
	duration    time.Duration
	callback    SchedulerCallback
}

// DeliverInfo.Call 调用计时任务回调函数
func (d *DeliverInfo) Call() {
	d.callback(d.duration, d.deliverTime)
}

type scheduleNode interface {
	getScheduleId() uint64
	start(scheduler *scheduler)
	stop()
}

type scheduler struct {
	maxScheduleId uint64
	schedulers    map[uint64]scheduleNode
	mutex         sync.Mutex
	deliver       chan *DeliverInfo
	context       context.Context
	waitGroup     *sync.WaitGroup
}

func NewScheduler() Scheduler {
	return &scheduler{
		maxScheduleId: 0,
		schedulers:    make(map[uint64]scheduleNode),
		deliver:       make(chan *DeliverInfo, 10),
	}
}

func (s *scheduler) Start(context context.Context, group *sync.WaitGroup) error {
	s.context = context
	s.waitGroup = group
	return nil
}

func (s *scheduler) Stop() {
	s.stopAllSchedulers()
}

func (s *scheduler) stopAllSchedulers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, schedule := range s.schedulers {
		schedule.stop()
	}
	s.schedulers = nil
}

func (s *scheduler) Done() <-chan *DeliverInfo {
	return s.deliver
}

func (s *scheduler) Schedule(duration time.Duration, isTicker bool, callback SchedulerCallback) (scheduleId uint64, err error) {
	var node scheduleNode
	scheduleId = atomic.AddUint64(&s.maxScheduleId, 1)
	if isTicker {
		node = newTickerNode(scheduleId, duration, callback)
	} else {
		node = newTimerNode(scheduleId, duration, callback)
	}
	if nil == node {
		return 0, fmt.Errorf("failed to create timer node")
	}

	s.mutex.Lock()
	s.schedulers[scheduleId] = node
	s.mutex.Unlock()

	node.start(s)
	return scheduleId, nil
}

func (s *scheduler) Unschedule(scheduleId uint64) error {
	s.mutex.Lock()
	schedule, ok := s.schedulers[scheduleId]
	if !ok {
		s.mutex.Unlock()
		return fmt.Errorf("schedule %v not found", scheduleId)
	}
	s.mutex.Unlock()
	schedule.stop()
	return nil
}

func (s *scheduler) addScheduleNode(node scheduleNode) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.schedulers[node.getScheduleId()] = node
}

func (s *scheduler) removeScheduleNode(scheduleId uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.schedulers, scheduleId)
}

// region: scheduleNodeBase
type scheduleRunner interface {
	done() <-chan time.Time
	stop()
	isFinish() bool
}

type scheduleBaseNode struct {
	scheduleId uint64
	duration   time.Duration
	callback   SchedulerCallback
	done       chan bool
	runner     scheduleRunner
}

func newScheduleNode(scheduleId uint64, duration time.Duration, callback SchedulerCallback, runner scheduleRunner) *scheduleBaseNode {
	return &scheduleBaseNode{
		scheduleId: scheduleId,
		duration:   duration,
		callback:   callback,
		done:       make(chan bool, 1),
		runner:     runner,
	}
}

func (n *scheduleBaseNode) getScheduleId() uint64 {
	return n.scheduleId
}

func (n *scheduleBaseNode) start(scheduler *scheduler) {
	scheduler.waitGroup.Add(1)
	go n.run(scheduler)
}

func (n *scheduleBaseNode) run(scheduler *scheduler) {
	defer scheduler.waitGroup.Done()
	defer scheduler.removeScheduleNode(n.scheduleId)
	defer logger.Debug("Scheduler|exit timer schedule %v routine", n.scheduleId)

	for {
		select {
		case <-scheduler.context.Done():
			logger.Debug("Scheduler|exit timer schedule %v when context is done.", n.scheduleId)
			return
		case <-n.done:
			logger.Debug("Scheduler|exit timer schedule %v when stop scheduler.", n.scheduleId)
			return
		case deliverTime := <-n.runner.done():
			logger.Debug("Scheduler|trigger timer schedule %v at %v", n.scheduleId, deliverTime)
			scheduler.deliver <- &DeliverInfo{
				scheduleId:  n.scheduleId,
				deliverTime: deliverTime,
				duration:    n.duration,
				callback:    n.callback,
			}
			if n.runner.isFinish() {
				return
			}
		}
	}
}

func (n *scheduleBaseNode) stop() {
	n.runner.stop()
	n.done <- true
}

// endregion: scheduleNodeBase

// region: timerRunner
type timerRunner struct {
	timer *time.Timer
}

func newTimerNode(scheduleId uint64, duration time.Duration, callback SchedulerCallback) *scheduleBaseNode {
	return newScheduleNode(scheduleId, duration, callback, &timerRunner{timer: time.NewTimer(duration)})
}

func (n *timerRunner) done() <-chan time.Time {
	return n.timer.C
}

func (n *timerRunner) stop() {
	n.timer.Stop()
}

func (n *timerRunner) isFinish() bool {
	return true
}

// endregion: timerRunner

// region: tickerRunner
type tickerRunner struct {
	ticker *time.Ticker
}

func newTickerNode(scheduleId uint64, duration time.Duration, callback SchedulerCallback) *scheduleBaseNode {
	return newScheduleNode(scheduleId, duration, callback, &tickerRunner{ticker: time.NewTicker(duration)})
}

func (n *tickerRunner) done() <-chan time.Time {
	return n.ticker.C
}

func (n *tickerRunner) stop() {
	n.ticker.Stop()
}

func (n *tickerRunner) isFinish() bool {
	return false
}

// endregion: tickerRunner

