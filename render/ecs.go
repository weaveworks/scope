package render

import (
	"github.com/weaveworks/scope/report"
)

// ECSTaskRenderer is a Renderer for Amazon ECS tasks.
var ECSTaskRenderer = MakeFilter(
	// TODO
	func(n report.Node) bool { return true },
	SelectECSTask,
)

// ECSServiceRenderer is a Renderer for Amazon ECS services.
var ECSServiceRenderer = MakeFilter(
	// TODO
	func(n report.Node) bool { return true },
	SelectECSService,
)
