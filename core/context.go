package core

import (
	"github.com/vulkan-go/vulkan"
)

// Vulkan context
type Context struct {
	surface                  vulkan.Surface        // Vulkan rendering surface (may be nil for headless mode)
	instance                 Instance              // Vulkan Instance
	physicalDevice           vulkan.PhysicalDevice // Physical device
	device                   vulkan.Device         // Logical vulkan device
	presentQueueFamilyIndex  QueueFamilyIndex      // Present queue family index
	graphicsQueueFamilyIndex QueueFamilyIndex      // Graphics queue family index
	computeQueueFamilyIndex  QueueFamilyIndex      // Compute queue family index
	transferQueueFamilyIndex QueueFamilyIndex      // Transfer queue family index
	graphicsCommandPool      vulkan.CommandPool    // Graphics command pool
	computeCommandPool       vulkan.CommandPool    // Compute command pool
	transferCommandPool      vulkan.CommandPool    // Transfer command pool
	presentQueue             vulkan.Queue          // Present queue
	graphicsQueue            vulkan.Queue          // Graphics queue
	computeQueue             vulkan.Queue          // Compute queue
	transferQueue            vulkan.Queue          // Transfer queue
}

// Creates vulkan context
func CreateContext(instance Instance, surface vulkan.Surface) (Context, error) {
	// First set the surface
	ctx := Context{}
	ctx.surface = surface
	ctx.instance = instance

	// Create physical device
	physicalDevice, err := PickPhysicalDevice(ctx.instance.instance)
	if err != nil {
		return ctx, err
	}
	ctx.physicalDevice = physicalDevice

	// Find queue families
	ctx.presentQueueFamilyIndex,
		ctx.graphicsQueueFamilyIndex,
		ctx.computeQueueFamilyIndex,
		ctx.transferQueueFamilyIndex = FindQueueFamilies(physicalDevice, surface)

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
	ctx.instance.Destroy()
}

func (ctx *Context) GetPhysicalDevice() vulkan.PhysicalDevice {
	return ctx.physicalDevice
}

func (ctx *Context) GetDevice() vulkan.Device {
	return ctx.device
}
