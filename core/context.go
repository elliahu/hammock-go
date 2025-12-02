package core

import (
	"github.com/vulkan-go/vulkan"
)

type Context struct {
	surface                  vulkan.Surface
	instance                 vulkan.Instance
	debugCallback            vulkan.DebugReportCallback
	physicalDevice           vulkan.PhysicalDevice
	device                   vulkan.Device
	presentQueueFamilyIndex  QueueFamilyIndex
	graphicsQueueFamilyIndex QueueFamilyIndex
	computeQueueFamilyIndex  QueueFamilyIndex
	transferQueueFamilyIndex QueueFamilyIndex
	graphicsCommandPool      vulkan.CommandPool
	computeCommandPool       vulkan.CommandPool
	transferCommandPool      vulkan.CommandPool
	presentQueue             vulkan.Queue
	graphicsQueue            vulkan.Queue
	computeQueue             vulkan.Queue
	transferQueue            vulkan.Queue
}

func CreateContext(surface vulkan.Surface) (*Context, error) {
	// First set the surface
	ctx := &Context{}
	ctx.surface = surface

	// Create Vulkan instance
	instance, err := CreateInstance()
	if err != nil {
		return ctx, err
	}

	ctx.instance = instance

	// Create debug messenger
	debugCallback, err := CreateDebugCallback(instance)
	if err != nil {
		return ctx, err
	}

	ctx.debugCallback = debugCallback

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

func (ctx *Context) Destroy() {
	DestroyCommandPools(ctx.device, ctx.graphicsCommandPool, ctx.computeCommandPool, ctx.transferCommandPool)
	DestroyDevice(ctx.device)
	DestroyDebugCallback(ctx.instance, ctx.debugCallback)
	DestroyInstance(ctx.instance)
}
