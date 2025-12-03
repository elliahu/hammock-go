package core

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/vulkan-go/vulkan"
)

type Instance struct {
	instance      vulkan.Instance
	debugCallback vulkan.DebugReportCallback
}

func (i *Instance) Instance() vulkan.Instance {
	return i.instance
}

// Debug callback function using the older DebugReport API
func debugCallback(
	flags vulkan.DebugReportFlags,
	objectType vulkan.DebugReportObjectType,
	object uint64,
	location uint,
	messageCode int32,
	pLayerPrefix string,
	pMessage string,
	pUserData unsafe.Pointer,
) vulkan.Bool32 {

	// Format severity
	var severity string
	if flags&vulkan.DebugReportFlags(vulkan.DebugReportErrorBit) != 0 {
		severity = "ERROR"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportWarningBit) != 0 {
		severity = "WARNING"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportPerformanceWarningBit) != 0 {
		severity = "PERFORMANCE"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportInformationBit) != 0 {
		severity = "INFO"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportDebugBit) != 0 {
		severity = "DEBUG"
	} else {
		severity = "UNKNOWN"
	}

	log.Printf("[%s] %s: %s\n", severity, pLayerPrefix, pMessage)

	// Return false to continue execution
	return vulkan.False
}

// Creates a debug callback object that is used as a messenger
func createDebugCallback(instance vulkan.Instance) (vulkan.DebugReportCallback, error) {
	createInfo := vulkan.DebugReportCallbackCreateInfo{
		SType: vulkan.StructureTypeDebugReportCallbackCreateInfo,
		Flags: vulkan.DebugReportFlags(
			vulkan.DebugReportErrorBit |
				vulkan.DebugReportWarningBit |
				vulkan.DebugReportPerformanceWarningBit,
		),
		PfnCallback: debugCallback,
	}

	var debugCallback vulkan.DebugReportCallback
	res := vulkan.CreateDebugReportCallback(instance, &createInfo, nil, &debugCallback)
	if res != vulkan.Success {
		return nil, fmt.Errorf("failed to create debug callback: %v", res)
	}

	return debugCallback, nil
}

// Destroys debug callback
func destroyDebugCallback(instance vulkan.Instance, callback vulkan.DebugReportCallback) {
	if callback != vulkan.NullDebugReportCallback {
		vulkan.DestroyDebugReportCallback(instance, callback, nil)
	}
}

// Creates Vulkan instance along with required instance extensions and layers.
// TODO make validation layers optional
// TODO use surface based on OS
func createInstance() (vulkan.Instance, error) {
	// Load Vulkan library
	if err := vulkan.SetDefaultGetInstanceProcAddr(); err != nil {
		return nil, fmt.Errorf("failed to load Vulkan: %w", err)
	}
	vulkan.Init()

	appName := "Hammock app\x00"
	engName := "HammockGo\x00"

	appInfo := vulkan.ApplicationInfo{
		SType:              vulkan.StructureTypeApplicationInfo,
		PApplicationName:   appName,
		ApplicationVersion: vulkan.MakeVersion(1, 0, 0),
		PEngineName:        engName,
		EngineVersion:      vulkan.MakeVersion(1, 0, 0),
		ApiVersion:         vulkan.ApiVersion10,
	}

	// Extensions in use
	extensions := []string{
		"VK_KHR_surface\x00",
		"VK_KHR_win32_surface\x00",
		"VK_EXT_debug_report\x00",
	}

	// Validation layers
	layers := []string{
		"VK_LAYER_KHRONOS_validation\x00",
	}

	instanceCreateInfo := vulkan.InstanceCreateInfo{
		SType:                   vulkan.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        &appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
	}

	// Create the actual instance
	var instance vulkan.Instance
	res := vulkan.CreateInstance(&instanceCreateInfo, nil, &instance)
	if res != vulkan.Success {
		return nil, fmt.Errorf("failed to create Vulkan instance: %v", res)
	}

	// This is needed on MoltenVk
	vulkan.InitInstance(instance)

	return instance, nil
}

// Destroy Vulkan instance
func destroyInstance(instance vulkan.Instance) {
	if instance != nil {
		vulkan.DestroyInstance(instance, nil)
	}
}

func CreateInstance() (Instance, error) {
	vulkanInstance := Instance{}

	// Create Vulkan instance
	instance, err := createInstance()
	if err != nil {
		return vulkanInstance, err
	}

	vulkanInstance.instance = instance

	// Create debug messenger
	debugCallback, err := createDebugCallback(instance)
	if err != nil {
		return vulkanInstance, err
	}

	vulkanInstance.debugCallback = debugCallback

	return vulkanInstance, nil
}

func (inst *Instance) Destroy() {
	destroyDebugCallback(inst.instance, inst.debugCallback)
	destroyInstance(inst.instance)
}
