package httpkit

import (
	"github.com/thedevsaddam/renderer"
)

var (
	// Render http response content
	Render = renderer.New()
)

// M describes handy type that represents data to send as response
type M map[string]interface{}
