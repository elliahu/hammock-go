package renderer

import "hammock-go/core"

type Renderer struct {
	context *core.Context // Vulkan context
}

func CreateRenderer(context *core.Context) Renderer {
	renderer := Renderer{}
	renderer.context = context

	return renderer
}
