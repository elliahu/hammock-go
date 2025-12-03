package core

import (
	"fmt"
	"math"

	"github.com/vulkan-go/vulkan"
)

type SwapChain struct {
	surfaceFormat vulkan.SurfaceFormat
	swapChain     vulkan.Swapchain
	device        vulkan.Device
	images        []vulkan.Image
	views         []vulkan.ImageView
}

func (sc *SwapChain) Create(
	instance vulkan.Instance,
	physicalDevice vulkan.PhysicalDevice,
	surface vulkan.Surface,
	device vulkan.Device,
	width uint32, height uint32,
	vSync bool) error {
	// Store old swapchain handle
	oldSwapChain := sc.swapChain
	sc.device = device

	// Get physical device surface properties and formats
	surfaceCaps := vulkan.SurfaceCapabilities{}
	res := vulkan.GetPhysicalDeviceSurfaceCapabilities(physicalDevice, surface, &surfaceCaps)
	if res != vulkan.Success {
		return fmt.Errorf("failed to get physical device surface capabilities")
	}

	// Explicitly copy C memory back to Go struct
	surfaceCaps.Deref()
	surfaceCaps.CurrentExtent.Deref()
	surfaceCaps.MinImageExtent.Deref()
	surfaceCaps.MaxImageExtent.Deref()

	var swapchainExtent vulkan.Extent2D
	// If width (and height) equals the special value 0xFFFFFFFF, the size of the surface will be set by the swapchain
	if surfaceCaps.CurrentExtent.Width == uint32(^uint32(0)) {
		swapchainExtent.Width = width
		swapchainExtent.Height = height
	} else {
		// If the surface size is defined, the swap chain size must match
		swapchainExtent = surfaceCaps.CurrentExtent
	}

	var presentModeCount uint32
	res = vulkan.GetPhysicalDeviceSurfacePresentModes(physicalDevice, surface, &presentModeCount, nil)
	if res != vulkan.Success || presentModeCount == 0 {
		return fmt.Errorf("failed to get present mode count")
	}

	presentModes := make([]vulkan.PresentMode, presentModeCount)
	res = vulkan.GetPhysicalDeviceSurfacePresentModes(physicalDevice, surface, &presentModeCount, presentModes)
	if res != vulkan.Success {
		return fmt.Errorf("failed to get present modes")
	}

	// The VK_PRESENT_MODE_FIFO_KHR mode must always be present as per spec
	// This mode waits for the vertical blank ("v-sync")
	var swapchainPresentMode vulkan.PresentMode = vulkan.PresentModeFifo

	// If v-sync is not requested, try to find a mailbox mode
	// It's the lowest latency non-tearing present mode available
	if !vSync {
		for i := range presentModeCount {
			if presentModes[i] == vulkan.PresentModeMailbox {
				swapchainPresentMode = vulkan.PresentModeMailbox
				break
			}

			if presentModes[i] == vulkan.PresentModeImmediate {
				swapchainPresentMode = vulkan.PresentModeImmediate
			}
		}
	}

	// Determine the number of images
	desiredNumberOfSwapchainImages := surfaceCaps.MinImageCount + 1
	if surfaceCaps.MaxImageCount > 0 && desiredNumberOfSwapchainImages > surfaceCaps.MaxImageCount {
		desiredNumberOfSwapchainImages = surfaceCaps.MaxImageCount
	}

	// Find the transformation of the surface
	var preTransform vulkan.SurfaceTransformFlagBits
	if (surfaceCaps.SupportedTransforms & vulkan.SurfaceTransformFlags(vulkan.SurfaceTransformIdentityBit)) != 0 {
		preTransform = vulkan.SurfaceTransformIdentityBit
	} else {
		preTransform = surfaceCaps.CurrentTransform
	}

	// Find a supported composite alpha format (not all devices support alpha opaque)
	compositeAlpha := vulkan.CompositeAlphaFlagBits(vulkan.CompositeAlphaOpaqueBit)
	// Simply select the first composite alpha format available
	compositeAlphaFlags := []vulkan.CompositeAlphaFlagBits{
		vulkan.CompositeAlphaOpaqueBit,
		vulkan.CompositeAlphaPreMultipliedBit,
		vulkan.CompositeAlphaPostMultipliedBit,
		vulkan.CompositeAlphaInheritBit,
	}
	for _, compositeAlphaFlag := range compositeAlphaFlags {
		if (surfaceCaps.SupportedCompositeAlpha & vulkan.CompositeAlphaFlags(compositeAlphaFlag)) != 0 {
			compositeAlpha = compositeAlphaFlag
			break
		}
	}

	// Find supported color format
	var formatCount uint32
	res = vulkan.GetPhysicalDeviceSurfaceFormats(physicalDevice, surface, &formatCount, nil)
	if res != vulkan.Success && formatCount == 0 {
		return fmt.Errorf("failed to get swapchain format count")
	}

	surfaceFormats := make([]vulkan.SurfaceFormat, formatCount)
	res = vulkan.GetPhysicalDeviceSurfaceFormats(physicalDevice, surface, &formatCount, surfaceFormats)
	if res != vulkan.Success && formatCount == 0 {
		return fmt.Errorf("failed to get swapchain formats")
	}

	selectedFormat := surfaceFormats[0]
	preferredImageFormats := []vulkan.Format{
		vulkan.FormatB8g8r8a8Unorm,
		vulkan.FormatR8g8b8a8Unorm,
		vulkan.FormatA8b8g8r8UnormPack32,
	}
	preferredMap := make(map[vulkan.Format]bool)
	for _, format := range preferredImageFormats {
		preferredMap[format] = true
	}
	for _, availableFormat := range surfaceFormats {
		if preferredMap[availableFormat.Format] {
			selectedFormat = availableFormat
			break
		}
	}

	sc.surfaceFormat = selectedFormat

	swapchainCreateInfo := vulkan.SwapchainCreateInfo{
		SType:                 vulkan.StructureTypeSwapchainCreateInfo,
		Surface:               surface,
		MinImageCount:         desiredNumberOfSwapchainImages,
		ImageFormat:           selectedFormat.Format,
		ImageColorSpace:       selectedFormat.ColorSpace,
		ImageExtent:           swapchainExtent,
		ImageArrayLayers:      1,
		ImageUsage:            vulkan.ImageUsageFlags(vulkan.ImageUsageColorAttachmentBit | vulkan.ImageUsageTransferSrcBit | vulkan.ImageUsageTransferDstBit),
		ImageSharingMode:      vulkan.SharingModeExclusive,
		QueueFamilyIndexCount: 0,
		PreTransform:          preTransform,
		CompositeAlpha:        compositeAlpha,
		PresentMode:           swapchainPresentMode,
		Clipped:               vulkan.True, // Setting clipped to VK_TRUE allows the implementation to discard rendering outside of the surface area
		OldSwapchain:          oldSwapChain,
	}

	var swapchainHandle vulkan.Swapchain

	res = vulkan.CreateSwapchain(device, &swapchainCreateInfo, nil, &swapchainHandle)
	if res != vulkan.Success {
		return fmt.Errorf("failed to create swapchain")
	}
	sc.swapChain = swapchainHandle

	// If an existing swap chain is re-created, destroy the old swap chain and the resources owned by the application (image views, images are owned by the swap chain)
	if oldSwapChain != vulkan.NullSwapchain {
		for i := range len(sc.images) {
			vulkan.DestroyImageView(device, sc.views[i], nil)
		}
		vulkan.DestroySwapchain(device, oldSwapChain, nil)
	}

	// Get the (new) swap chain images
	imageCount := uint32(len(sc.images))
	res = vulkan.GetSwapchainImages(device, sc.swapChain, &imageCount, nil)
	if res != vulkan.Success {
		return fmt.Errorf("failed to get swapchain image count")
	}
	newImages := make([]vulkan.Image, imageCount)
	copy(newImages, sc.images)
	sc.images = newImages
	res = vulkan.GetSwapchainImages(device, sc.swapChain, &imageCount, sc.images)
	if res != vulkan.Success {
		return fmt.Errorf("failed to get swapchain image count")
	}

	// Get the swap chain buffers containing the image and imageview
	newImageViews := make([]vulkan.ImageView, imageCount)
	copy(newImageViews, sc.views)
	sc.views = newImageViews
	for i := range len(sc.images) {
		colorAttachmentView := vulkan.ImageViewCreateInfo{
			SType:      vulkan.StructureTypeImageViewCreateInfo,
			Image:      sc.images[i],
			ViewType:   vulkan.ImageViewType2d,
			Format:     sc.surfaceFormat.Format,
			Components: vulkan.ComponentMapping{R: vulkan.ComponentSwizzleR, G: vulkan.ComponentSwizzleG, B: vulkan.ComponentSwizzleB, A: vulkan.ComponentSwizzleA},
			SubresourceRange: vulkan.ImageSubresourceRange{
				AspectMask:     vulkan.ImageAspectFlags(vulkan.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}

		res = vulkan.CreateImageView(device, &colorAttachmentView, nil, &sc.views[i])
		if res != vulkan.Success {
			return fmt.Errorf("failed to create swapchain image view")
		}
	}

	return nil
}

func (sc *SwapChain) Destroy() {
	if sc.swapChain != vulkan.NullSwapchain {
		for i := range len(sc.images) {
			vulkan.DestroyImageView(sc.device, sc.views[i], nil)
		}
		vulkan.DestroySwapchain(sc.device, sc.swapChain, nil)
	}

	sc.swapChain = vulkan.NullSwapchain
}

func (sc *SwapChain) AcquireNextImage(presentCompleteSemaphore vulkan.Semaphore, imageIndex *uint32) vulkan.Result {
	// By setting timeout to UINT64_MAX we will always wait until the next image has been acquired or an actual error is thrown
	// With that we don't have to handle VK_NOT_READY
	return vulkan.AcquireNextImage(sc.device, sc.swapChain, math.MaxUint64, presentCompleteSemaphore, nil, imageIndex)
}
