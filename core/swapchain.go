package core

import (
	"fmt"
	"math"

	"github.com/bbredesen/go-vk"
)

type SwapChain struct {
	surfaceFormat vk.SurfaceFormatKHR
	swapChain     vk.SwapchainKHR
	device        vk.Device
	images        []vk.Image
	views         []vk.ImageView
}

func (sc *SwapChain) Create(
	instance vk.Instance,
	physicalDevice vk.PhysicalDevice,
	surface vk.SurfaceKHR,
	device vk.Device,
	width uint32, height uint32,
	vSync bool) error {
	// Store old swapchain handle
	oldSwapChain := sc.swapChain
	sc.device = device

	// Get physical device surface properties and formats
	surfaceCaps, err := vk.GetPhysicalDeviceSurfaceCapabilitiesKHR(physicalDevice, surface)
	if err != nil {
		return fmt.Errorf("failed to get physical device surface capabilities")
	}

	var swapchainExtent vk.Extent2D
	// If width (and height) equals the special value 0xFFFFFFFF, the size of the surface will be set by the swapchain
	if surfaceCaps.CurrentExtent.Width == uint32(^uint32(0)) {
		swapchainExtent.Width = width
		swapchainExtent.Height = height
	} else {
		// If the surface size is defined, the swap chain size must match
		swapchainExtent = surfaceCaps.CurrentExtent
	}

	presentModes, err := vk.GetPhysicalDeviceSurfacePresentModesKHR(physicalDevice, surface)
	if err != nil {
		return fmt.Errorf("failed to get physical device surface present modes")
	}

	// The VK_PRESENT_MODE_FIFO_KHR mode must always be present as per spec
	// This mode waits for the vertical blank ("v-sync")
	var swapchainPresentMode vk.PresentModeKHR = vk.PRESENT_MODE_FIFO_KHR

	// If v-sync is not requested, try to find a mailbox mode
	// It's the lowest latency non-tearing present mode available
	if !vSync {
		for i := range len(presentModes) {
			if presentModes[i] == vk.PRESENT_MODE_MAILBOX_KHR {
				swapchainPresentMode = vk.PRESENT_MODE_MAILBOX_KHR
				break
			}

			if presentModes[i] == vk.PRESENT_MODE_IMMEDIATE_KHR {
				swapchainPresentMode = vk.PRESENT_MODE_IMMEDIATE_KHR
			}
		}
	}

	// Determine the number of images
	desiredNumberOfSwapchainImages := surfaceCaps.MinImageCount + 1
	if surfaceCaps.MaxImageCount > 0 && desiredNumberOfSwapchainImages > surfaceCaps.MaxImageCount {
		desiredNumberOfSwapchainImages = surfaceCaps.MaxImageCount
	}

	// Find the transformation of the surface
	var preTransform vk.SurfaceTransformFlagBitsKHR
	if (surfaceCaps.SupportedTransforms & vk.SURFACE_TRANSFORM_IDENTITY_BIT_KHR) != 0 {
		preTransform = vk.SURFACE_TRANSFORM_IDENTITY_BIT_KHR
	} else {
		preTransform = surfaceCaps.CurrentTransform
	}

	// Find a supported composite alpha format (not all devices support alpha opaque)
	compositeAlpha := vk.COMPOSITE_ALPHA_OPAQUE_BIT_KHR
	// Simply select the first composite alpha format available
	compositeAlphaFlags := []vk.CompositeAlphaFlagBitsKHR{
		vk.COMPOSITE_ALPHA_OPAQUE_BIT_KHR,
		vk.COMPOSITE_ALPHA_PRE_MULTIPLIED_BIT_KHR,
		vk.COMPOSITE_ALPHA_POST_MULTIPLIED_BIT_KHR,
		vk.COMPOSITE_ALPHA_INHERIT_BIT_KHR,
	}
	for _, compositeAlphaFlag := range compositeAlphaFlags {
		if (surfaceCaps.SupportedCompositeAlpha & compositeAlphaFlag) != 0 {
			compositeAlpha = compositeAlphaFlag
			break
		}
	}

	// Find supported color format
	surfaceFormats, err := vk.GetPhysicalDeviceSurfaceFormatsKHR(physicalDevice, surface)
	if err != nil {
		return fmt.Errorf("failed to get swapchain formats")
	}

	selectedFormat := surfaceFormats[0]
	preferredImageFormats := []vk.Format{
		vk.FORMAT_B8G8R8A8_UNORM,
		vk.FORMAT_R8G8B8A8_UNORM,
		vk.FORMAT_A8B8G8R8_SRGB_PACK32,
	}
	preferredMap := make(map[vk.Format]bool)
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

	swapchainCreateInfo := vk.SwapchainCreateInfoKHR{
		Surface:          surface,
		MinImageCount:    desiredNumberOfSwapchainImages,
		ImageFormat:      selectedFormat.Format,
		ImageColorSpace:  selectedFormat.ColorSpace,
		ImageExtent:      swapchainExtent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.IMAGE_USAGE_COLOR_ATTACHMENT_BIT | vk.IMAGE_USAGE_TRANSFER_SRC_BIT | vk.IMAGE_USAGE_TRANSFER_DST_BIT,
		ImageSharingMode: vk.SHARING_MODE_EXCLUSIVE,
		PreTransform:     preTransform,
		CompositeAlpha:   compositeAlpha,
		PresentMode:      swapchainPresentMode,
		Clipped:          true, // Setting clipped to VK_TRUE allows the implementation to discard rendering outside of the surface area
		OldSwapchain:     oldSwapChain,
	}

	swapchainHandle, err := vk.CreateSwapchainKHR(device, &swapchainCreateInfo, nil)
	if err != nil {
		return fmt.Errorf("failed to create swapchain")
	}
	sc.swapChain = swapchainHandle

	// If an existing swap chain is re-created, destroy the old swap chain and the resources owned by the application (image views, images are owned by the swap chain)
	if oldSwapChain != vk.SwapchainKHR(vk.NULL_HANDLE) {
		for i := range len(sc.images) {
			vk.DestroyImageView(device, sc.views[i], nil)
		}
		vk.DestroySwapchainKHR(device, oldSwapChain, nil)
	}

	swapchainImages, err := vk.GetSwapchainImagesKHR(device, sc.swapChain)
	if err != nil {
		return fmt.Errorf("failed to get swapchain images")
	}
	sc.images = swapchainImages

	// Get the swap chain buffers containing the image and imageview
	newImageViews := make([]vk.ImageView, len(sc.images))
	copy(newImageViews, sc.views)
	sc.views = newImageViews
	for i := range len(sc.images) {
		colorAttachmentView := vk.ImageViewCreateInfo{
			Image:    sc.images[i],
			ViewType: vk.IMAGE_VIEW_TYPE_2D,
			Format:   sc.surfaceFormat.Format,
			Components: vk.ComponentMapping{
				R: vk.COMPONENT_SWIZZLE_R,
				G: vk.COMPONENT_SWIZZLE_G,
				B: vk.COMPONENT_SWIZZLE_B,
				A: vk.COMPONENT_SWIZZLE_A,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.IMAGE_ASPECT_COLOR_BIT,
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}

		imageView, err := vk.CreateImageView(device, &colorAttachmentView, nil)
		if err != nil {
			return fmt.Errorf("failed to create swapchain image view")
		}
		sc.views[i] = imageView
	}

	return nil
}

func (sc *SwapChain) Destroy() {
	if sc.swapChain != vk.SwapchainKHR(vk.NULL_HANDLE) {
		for i := range len(sc.images) {
			vk.DestroyImageView(sc.device, sc.views[i], nil)
		}
		vk.DestroySwapchainKHR(sc.device, sc.swapChain, nil)
	}

	sc.swapChain = vk.SwapchainKHR(vk.NULL_HANDLE)
}

func (sc *SwapChain) AcquireNextImage(presentCompleteSemaphore vk.Semaphore) (uint32, error) {
	// By setting timeout to UINT64_MAX we will always wait until the next image has been acquired or an actual error is thrown
	// With that we don't have to handle VK_NOT_READY
	return vk.AcquireNextImageKHR(sc.device, sc.swapChain, math.MaxUint64, presentCompleteSemaphore, vk.Fence(vk.NULL_HANDLE))
}
