package container

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Service interface {
	// Start 启动服务 routine
	Start(ctx context.Context, wg *sync.WaitGroup) error
	// Stop 通知服务停止
	Stop()
}

type Container struct {
	wg			sync.WaitGroup
	ctx			context.Context
	cancel		context.CancelFunc
	
	services	[]Service
}

func NewContainer() *Container {
	ctx, cancel := context.WithCancel(context.Background())
	return &Container{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Container) AddService(service Service) {
	c.services = append(c.services, service)
}

func (c *Container) Run() error {
	for _, service := range c.services {
		if err := service.Start(c.ctx, &c.wg); nil != err {
			return err
		}
	}
	c.waitForDone()

	for _, service := range c.services {
		service.Stop()
	}
	c.cancel()
	c.wg.Wait()
	return nil
}

func (c *Container) waitForDone() {
	quitCh := make(chan struct{}, 1)

	// 系统信号
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT)
		for {
			msg := <-ch
			if msg == syscall.SIGINT {
				quitCh <- struct{}{}
				return
			}
		}
	}()

	<-quitCh // 等待结束
}

