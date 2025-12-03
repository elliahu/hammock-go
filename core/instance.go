package core

import (
	"fmt"

	"github.com/bbredesen/go-vk"
)

// Creates Vulkan instance along with required instance extensions and layers.
// TODO make validation layers optional
// TODO use surface based on OS
// TODO create debug callback
func CreateInstance() (vk.Instance, error) {
	appName := "Hammock app"
	engName := "HammockGo"

	appInfo := vk.ApplicationInfo{
		PApplicationName:   appName,
		ApplicationVersion: vk.MAKE_VERSION(0, 0, 1),
		PEngineName:        engName,
		EngineVersion:      vk.MAKE_VERSION(0, 0, 1),
		ApiVersion:         vk.MAKE_VERSION(1, 3, 0),
	}

	// Extensions in use
	extensions := []string{
		vk.KHR_SURFACE_EXTENSION_NAME,
		vk.KHR_WIN32_SURFACE_EXTENSION_NAME,
	}

	// Validation layers
	layers := []string{
		"VK_LAYER_KHRONOS_validation",
	}

	instanceCreateInfo := vk.InstanceCreateInfo{
		PApplicationInfo:        &appInfo,
		PpEnabledExtensionNames: extensions,
		PpEnabledLayerNames:     layers,
	}

	// Create the actual instance
	instance, err := vk.CreateInstance(&instanceCreateInfo, nil)
	if err != nil {
		return vk.Instance(vk.NULL_HANDLE), fmt.Errorf("failed to create Vulkan instance")
	}

	return instance, nil
}

// Destroy Vulkan instance
func DestroyInstance(instance vk.Instance) {
	if instance != vk.Instance(vk.NULL_HANDLE) {
		vk.DestroyInstance(instance, nil)
	}
}
