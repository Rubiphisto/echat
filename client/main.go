package main

import (
	"echat/client/console"
	"echat/client/session"
	"echat/utils/container"
	"echat/utils/logger"
)

func addService(c *container.Container, service container.Service) {
	if nil == service {
		return
	}
	c.AddService(service)
}

func main() {
	c := container.NewContainer()
	addService(c, session.NewSession())
	addService(c, console.NewConsole())
	if err := c.Run(); nil != err {
		logger.Error("Failed to start the container with error %v", err)
		return
	}
}

