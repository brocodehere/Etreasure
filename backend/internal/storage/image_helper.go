package storage

import (
	"strings"
)

// ImageURLHelper provides centralized image URL formatting
type ImageURLHelper struct {
	r2Client *R2Client
}

// NewImageURLHelper creates a new image URL helper
func NewImageURLHelper(r2Client *R2Client) *ImageURLHelper {
	return &ImageURLHelper{r2Client: r2Client}
}

// FormatImageURL converts various image path formats to canonical public URLs
// Handles: R2 keys (product/uuid.webp), local paths (/uploads/), and full URLs
func (h *ImageURLHelper) FormatImageURL(imagePath *string) *string {
	if imagePath == nil || h.r2Client == nil {
		return imagePath
	}

	path := *imagePath
	if path == "" {
		return nil
	}

	// If it's an R2 key (starts with product/, banner/, category/), use R2 public URL
	if strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/") {
		url := h.r2Client.PublicURL(path)
		return &url
	}

	// If it's a local path starting with /uploads/, convert to full URL
	if strings.HasPrefix(path, "/uploads/") {
		// In production, this should use the actual frontend domain
		// For now, we'll treat it as a relative path that the frontend will handle
		return imagePath
	}

	// If it's already a full URL, keep as is
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return imagePath
	}

	// Otherwise, assume it's an R2 key and convert to public URL
	url := h.r2Client.PublicURL(path)
	return &url
}

// GetImageKeyAndURL returns both the original key and the public URL
// Useful for API responses that need both fields
func (h *ImageURLHelper) GetImageKeyAndURL(imagePath *string) (key *string, url *string) {
	if imagePath == nil {
		return nil, nil
	}

	path := *imagePath
	if path == "" {
		return nil, nil
	}

	// For R2 keys, return the key as-is and the public URL
	if strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/") {
		publicURL := h.r2Client.PublicURL(path)
		return imagePath, &publicURL
	}

	// For local paths, return the path and nil URL (frontend will handle)
	if strings.HasPrefix(path, "/uploads/") {
		return imagePath, nil
	}

	// For full URLs, return nil key and the URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return nil, imagePath
	}

	// Otherwise, treat as R2 key
	publicURL := h.r2Client.PublicURL(path)
	return imagePath, &publicURL
}

// GetFallbackImageURL returns a fallback image URL based on the context
func (h *ImageURLHelper) GetFallbackImageURL(context string) string {
	// Use R2 public base URL for fallback images
	baseURL := strings.TrimSuffix(h.r2Client.PublicURL(""), "/")

	switch context {
	case "product":
		return baseURL + "/product-placeholder.webp"
	case "banner":
		return baseURL + "/banner-placeholder.webp"
	case "category":
		return baseURL + "/category-placeholder.webp"
	default:
		return baseURL + "/placeholder.webp"
	}
}

// FormatImageURLWithFallback returns the formatted image URL or a fallback if none exists
func (h *ImageURLHelper) FormatImageURLWithFallback(imagePath *string, context string) string {
	if formattedURL := h.FormatImageURL(imagePath); formattedURL != nil && *formattedURL != "" {
		return *formattedURL
	}

	// Return fallback URL
	return h.GetFallbackImageURL(context)
}
