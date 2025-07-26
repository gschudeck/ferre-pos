// Package utils proporciona utilidades de red con notación húngara
package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// extractClientIP extrae la IP real del cliente considerando proxies y load balancers
func ExtractClientIP(ptrCtx *gin.Context) string {
	// Verificar headers de proxy en orden de prioridad
	arrHeadersToCheck := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Akamai
	}

	for _, strHeader := range arrHeadersToCheck {
		strIP := ptrCtx.GetHeader(strHeader)
		if strIP != "" {
			// X-Forwarded-For puede contener múltiples IPs separadas por comas
			if strHeader == "X-Forwarded-For" {
				arrIPs := strings.Split(strIP, ",")
				if len(arrIPs) > 0 {
					strIP = strings.TrimSpace(arrIPs[0])
				}
			}

			// Validar que sea una IP válida
			if isValidIP(strIP) {
				return strIP
			}
		}
	}

	// Fallback a RemoteAddr
	strRemoteAddr := ptrCtx.Request.RemoteAddr
	strIP, _, _ := net.SplitHostPort(strRemoteAddr)

	if isValidIP(strIP) {
		return strIP
	}

	// Último fallback
	return "unknown"
}

// isValidIP verifica si una string es una IP válida
func isValidIP(strIP string) bool {
	if strIP == "" {
		return false
	}

	// Remover espacios
	strIP = strings.TrimSpace(strIP)

	// Verificar que no sea una IP privada o reservada en contexto público
	ptrParsedIP := net.ParseIP(strIP)
	if ptrParsedIP == nil {
		return false
	}

	// Verificar que no sea localhost
	if ptrParsedIP.IsLoopback() {
		return false
	}

	return true
}

// IsPrivateIP verifica si una IP es privada
func IsPrivateIP(strIP string) bool {
	ptrParsedIP := net.ParseIP(strIP)
	if ptrParsedIP == nil {
		return false
	}

	return ptrParsedIP.IsPrivate()
}

// IsLocalhost verifica si una IP es localhost
func IsLocalhost(strIP string) bool {
	ptrParsedIP := net.ParseIP(strIP)
	if ptrParsedIP == nil {
		return false
	}

	return ptrParsedIP.IsLoopback()
}
