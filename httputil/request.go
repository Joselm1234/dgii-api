package httputil

import (
	"net/http"
	"strings"
)

func GetUserIP(r *http.Request) string {
	// Get the IP address from the request's RemoteAddr field
	ip := r.RemoteAddr

	// If the IP address contains a port number, remove it
	if index := strings.IndexByte(ip, ':'); index >= 0 {
		ip = ip[:index]
	}

	// Return the extracted IP address
	return ip
}

func GetUserBrowserAndOS(r *http.Request) (browser, os string) {
	// Get the User-Agent header from the request
	userAgent := r.Header.Get("User-Agent")

	// Extract the browser from the User-Agent string
	if strings.Contains(userAgent, "MSIE") || strings.Contains(userAgent, "Trident") {
		browser = "Internet Explorer"
	} else if strings.Contains(userAgent, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "Chrome") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "Safari") {
		browser = "Safari"
	} else if strings.Contains(userAgent, "OPR") {
		browser = "Opera"
	} else {
		browser = "Unknown"
	}

	// Extract the OS from the User-Agent string
	if strings.Contains(userAgent, "Windows") {
		os = "Windows"
	} else if strings.Contains(userAgent, "Macintosh") {
		os = "MacOS"
	} else if strings.Contains(userAgent, "Linux") {
		os = "Linux"
	} else if strings.Contains(userAgent, "Android") {
		os = "Android"
	} else if strings.Contains(userAgent, "iOS") {
		os = "iOS"
	} else {
		os = "Unknown"
	}

	return browser, os
}
