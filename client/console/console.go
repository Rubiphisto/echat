package console

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
)

var (
	console Console
)

type CmdHandler func(params []string)

type Console struct {
	context			context.Context
	cancel			context.CancelFunc
	handlers		map[string]CmdHandler
	maxHandlerId	uint32
}

func init() {
	console.handlers = make(map[string]CmdHandler)
}

func NewConsole() *Console {
	return &console
}

func (c *Console) Start(ctx context.Context, wg *sync.WaitGroup) error {
	c.context, c.cancel = context.WithCancel(ctx)
	c.run(wg)
	return nil
}

func (c *Console) Stop() {
}

func (c *Console) AddHandler(cmd string, handler CmdHandler) {
	c.handlers[cmd] = handler
}

func (c *Console) DelHandler(cmd string) {
	delete(c.handlers, cmd)
}

func (c *Console) run(wg *sync.WaitGroup) {
	wg.Add(1)
	
	lines := make(chan string)
	
	go func() {
		defer wg.Done()
		
		for {
			select {
			case <-c.context.Done():
				return
			case line := <-lines:
				cmds := strings.Split(line, " ")
				if 0 == len(cmds) {
					break
				}
				handler, ok := c.handlers[cmds[0]]
				if !ok {
					break
				}
				handler(cmds[1:])
				break
			}
		}
	} ()
	
	go func() {
		reader := bufio.NewReader(os.Stdin)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	} ()
}
