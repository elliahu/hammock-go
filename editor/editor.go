package editor

import (
	"fmt"
	"hammock-go/core"
	"hammock-go/renderer"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/vulkan-go/vulkan"
)

type Editor struct {
	window   *glfw.Window
	surface  vulkan.Surface
	instance *core.Instance
	context  *core.Context
	renderer *renderer.Renderer
}

func (edit *Editor) mainLoop() {

}

func createWindowAndSurface(editor *Editor) error {
	// Initialize GLFW
	err := glfw.Init()
	if err != nil {
		return fmt.Errorf("failed to init GLFW")
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)

	// Create window
	editor.window, err = glfw.CreateWindow(1920, 1080, "HammockGO Editor", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create window")
	}

	// Create surface
	surface, err := editor.window.CreateWindowSurface(editor.instance.Instance(), nil)
	if err != nil {
		return fmt.Errorf("failed to create surface")
	}
	editor.surface = vulkan.SurfaceFromPointer(surface)

	return nil
}

func CreateEditor() (*Editor, error) {
	editor := &Editor{}

	// Create instance
	instance, err := core.CreateInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to create instance")
	}
	editor.instance = instance

	// Create window and surface
	createWindowAndSurface(editor)

	// Create context
	core.CreateContext(editor.instance, editor.surface)

	// Create renderer
	editor.renderer = renderer.CreateRenderer(editor.context)

	return editor, nil
}

func (edit *Editor) Run() {
	for !edit.window.ShouldClose() {
		edit.mainLoop()
		glfw.PollEvents()
	}
}

func (edit *Editor) Destroy() {
	glfw.Terminate()
}
