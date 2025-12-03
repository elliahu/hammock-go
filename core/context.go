package core

import (
	"fmt"

	"github.com/bbredesen/go-vk"
)

// Vulkan context
type Context struct {
	surface                  vk.SurfaceKHR     // Vulkan rendering surface (may be nil for headless mode)
	instance                 vk.Instance       // Vulkan Instance
	physicalDevice           vk.PhysicalDevice // Physical device
	device                   vk.Device         // Logical vulkan device
	presentQueueFamilyIndex  QueueFamilyIndex  // Present queue family index
	graphicsQueueFamilyIndex QueueFamilyIndex  // Graphics queue family index
	computeQueueFamilyIndex  QueueFamilyIndex  // Compute queue family index
	transferQueueFamilyIndex QueueFamilyIndex  // Transfer queue family index
	graphicsCommandPool      vk.CommandPool    // Graphics command pool
	computeCommandPool       vk.CommandPool    // Compute command pool
	transferCommandPool      vk.CommandPool    // Transfer command pool
	presentQueue             vk.Queue          // Present queue
	graphicsQueue            vk.Queue          // Graphics queue
	computeQueue             vk.Queue          // Compute queue
	transferQueue            vk.Queue          // Transfer queue
}

// Creates vulkan context
func CreateContext(instance vk.Instance, surface vk.SurfaceKHR) (Context, error) {
	// First set the surface
	ctx := Context{}
	ctx.surface = surface
	ctx.instance = instance

	// Create physical device
	physicalDevice, err := PickPhysicalDevice(instance)
	if err != nil {
		return ctx, err
	}
	ctx.physicalDevice = physicalDevice

	// Find queue families
	ctx.presentQueueFamilyIndex,
		ctx.graphicsQueueFamilyIndex,
		ctx.computeQueueFamilyIndex,
		ctx.transferQueueFamilyIndex, err = FindQueueFamilies(physicalDevice, surface)
	if err != nil {
		return ctx, fmt.Errorf("failed to find queue families")
	}

	// Create logical device
	device, presentQueue, graphicsQueue, computeQueue, transferQueue, err := CreateDevice(physicalDevice, ctx.presentQueueFamilyIndex,
		ctx.graphicsQueueFamilyIndex, ctx.computeQueueFamilyIndex, ctx.transferQueueFamilyIndex)
	if err != nil {
		return ctx, err
	}

	ctx.device = device
	ctx.presentQueue = presentQueue
	ctx.graphicsQueue = graphicsQueue
	ctx.computeQueue = computeQueue
	ctx.transferQueue = transferQueue

	// Create command pools
	graphicsCommandPool, computeCommandPool, transferCommandPool, err := CreateCommandPools(device, ctx.graphicsQueueFamilyIndex,
		ctx.computeQueueFamilyIndex, ctx.transferQueueFamilyIndex)
	if err != nil {
		return ctx, err
	}

	ctx.graphicsCommandPool = graphicsCommandPool
	ctx.computeCommandPool = computeCommandPool
	ctx.transferCommandPool = transferCommandPool

	return ctx, nil
}

// Destroy Vulkan context
func (ctx *Context) Destroy() {
	DestroyCommandPools(ctx.device, ctx.graphicsCommandPool, ctx.computeCommandPool, ctx.transferCommandPool)
	DestroyDevice(ctx.device)
	DestroyInstance(ctx.instance)
}

func (ctx *Context) GetPhysicalDevice() vk.PhysicalDevice {
	return ctx.physicalDevice
}

func (ctx *Context) GetDevice() vk.Device {
	return ctx.device
}
