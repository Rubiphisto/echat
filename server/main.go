package main

import (
	"echat/server/sessions"
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
	addService(c, sessions.GetSessionManager())
	if err := c.Run(); nil != err {
		logger.Error("Failed to start the container with error %v", err)
		return
	}
}

