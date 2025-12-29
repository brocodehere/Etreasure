package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PreviewRequest struct {
	Type    string                 `json:"type" binding:"required"` // product | banner | category | offer
	Content map[string]interface{} `json:"content" binding:"required"`
}

type PreviewResponse struct {
	HTML string `json:"html"`
}

func (h *Handler) Preview(c *gin.Context) {
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var html string
	switch req.Type {
	case "product":
		html = renderProductPreview(req.Content)
	case "banner":
		html = renderBannerPreview(req.Content)
	case "category":
		html = renderCategoryPreview(req.Content)
	case "offer":
		html = renderOfferPreview(req.Content)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preview type"})
		return
	}

	c.JSON(http.StatusOK, PreviewResponse{HTML: html})
}

func renderProductPreview(data map[string]interface{}) string {
	title := asString(data["title"])
	description := asString(data["description"])
	price := asFloat(data["price"])
	currency := asString(data["currency"])
	image := asString(data["image_url"])

	var imageHtml string
	if image != "" {
		imageHtml = `<img class="w-full h-48 object-cover" src="` + image + `" alt="` + title + `">`
	} else {
		imageHtml = `<div class="w-full h-48 bg-gray-200"></div>`
	}

	return `<div class="max-w-sm rounded overflow-hidden shadow-lg bg-white">` +
		imageHtml +
		`<div class="px-6 py-4">
			<div class="font-bold text-xl mb-2">` + title + `</div>
			<p class="text-gray-700 text-base">` + description + `</p>
		</div>
		<div class="px-6 pt-4 pb-2">
			<span class="inline-block bg-gold rounded-full px-3 py-1 text-sm font-semibold text-white mr-2">
				` + currency + ` ` + strconv.FormatFloat(price, 'f', 2, 64) + `
			</span>
		</div>
	</div>`
}

func renderBannerPreview(data map[string]interface{}) string {
	title := asString(data["title"])
	image := asString(data["image_url"])
	link := asString(data["link_url"])

	var imageHtml string
	if image != "" {
		imageHtml = `<img class="absolute inset-0 w-full h-full object-cover opacity-50" src="` + image + `" alt="` + title + `">`
	}
	var linkHtml string
	if link != "" {
		linkHtml = `<a href="` + link + `" class="bg-gold text-white px-6 py-2 rounded hover:bg-yellow-600 transition">Shop Now</a>`
	}

	return `<div class="relative w-full h-64 bg-gradient-to-r from-maroon to-red-800 rounded-lg overflow-hidden">` +
		imageHtml +
		`<div class="relative z-10 flex flex-col justify-center items-center h-full text-white text-center px-6">
			<h2 class="text-3xl font-bold mb-2">` + title + `</h2>` +
		linkHtml +
		`</div>
	</div>`
}

func renderCategoryPreview(data map[string]interface{}) string {
	name := asString(data["name"])
	description := asString(data["description"])
	image := asString(data["image_url"])

	var imageHtml string
	if image != "" {
		imageHtml = `<img class="w-full h-48 object-cover" src="` + image + `" alt="` + name + `">`
	} else {
		imageHtml = `<div class="w-full h-48 bg-gray-200"></div>`
	}

	return `<div class="max-w-sm rounded overflow-hidden shadow-lg bg-white">` +
		imageHtml +
		`<div class="px-6 py-4">
			<div class="font-bold text-xl mb-2">` + name + `</div>
			<p class="text-gray-700 text-base">` + description + `</p>
		</div>
		<div class="px-6 pt-4 pb-2">
			<span class="inline-block bg-maroon text-white rounded-full px-3 py-1 text-sm font-semibold">
				View Products
			</span>
		</div>
	</div>`
}

func renderOfferPreview(data map[string]interface{}) string {
	title := asString(data["title"])
	description := asString(data["description"])
	discountType := asString(data["discount_type"])
	discountValue := asFloat(data["discount_value"])
	image := asString(data["image_url"])

	discountText := ""
	if discountType == "percentage" {
		discountText = strconv.FormatFloat(discountValue, 'f', 0, 64) + "% OFF"
	} else {
		discountText = "$" + strconv.FormatFloat(discountValue, 'f', 2, 64) + " OFF"
	}

	var imageHtml string
	if image != "" {
		imageHtml = `<img class="absolute inset-0 w-full h-full object-cover opacity-40" src="` + image + `" alt="` + title + `">`
	}
	var descriptionHtml string
	if description != "" {
		descriptionHtml = `<p class="text-sm">` + description + `</p>`
	}

	return `<div class="relative w-full h-48 bg-gradient-to-r from-green-600 to-emerald-700 rounded-lg overflow-hidden">` +
		imageHtml +
		`<div class="relative z-10 flex flex-col justify-center items-center h-full text-white text-center px-6">
			<span class="bg-yellow-400 text-black px-3 py-1 rounded-full text-sm font-bold mb-2">` + discountText + `</span>
			<h3 class="text-2xl font-bold mb-2">` + title + `</h3>` +
		descriptionHtml +
		`</div>
	</div>`
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	if s, ok := v.(map[string]interface{}); ok {
		if str, ok := s["s"].(string); ok {
			return str
		}
	}
	return ""
}

func asFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	if f, ok := v.(map[string]interface{}); ok {
		if f64, ok := f["f"].(float64); ok {
			return f64
		}
	}
	return 0
}
