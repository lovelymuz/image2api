package chatgpt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL                  = "https://chatgpt.com"
	defaultUserAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36 Edg/149.0.0.0"
	defaultClientVersion     = "prod-ab8a6348980a3e1d771c463b9f4f3e4e584f2769"
	defaultClientBuildNumber = "7624276"
	defaultPOWScript         = "https://chatgpt.com/backend-api/sentinel/sdk.js"
)

var (
	fileServiceIDPattern = regexp.MustCompile(`file-service://([A-Za-z0-9_-]+)`)
	sedimentIDPattern    = regexp.MustCompile(`sediment://([A-Za-z0-9_-]+)`)
	realImageIDPattern   = regexp.MustCompile(`\bfile_00000000[a-f0-9]{24}\b`)
	conversationIDRE     = regexp.MustCompile(`"conversation_id"\s*:\s*"([^"]+)"`)
	scriptSrcRE          = regexp.MustCompile(`<script[^>]+src="([^"]+)"`)
	dataBuildPathRE      = regexp.MustCompile(`c/[^/]*/_`)
	htmlDataBuildRE      = regexp.MustCompile(`<html[^>]*data-build="([^"]*)"`)
)

func stringValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

func intValue(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case float32:
		return int(x)
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}

func decodeJWTPayload(token string) map[string]any {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) < 2 {
		return map[string]any{}
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func newUUID() string {
	return uuid.NewString()
}

func clip(v []byte, n int) string {
	s := strings.TrimSpace(string(v))
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func parsePOWResources(html string) ([]string, string) {
	matches := scriptSrcRE.FindAllStringSubmatch(html, -1)
	sources := make([]string, 0, len(matches))
	dataBuild := ""
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		src := strings.TrimSpace(match[1])
		if src == "" {
			continue
		}
		sources = append(sources, src)
		if dataBuild == "" {
			if path := dataBuildPathRE.FindString(src); path != "" {
				dataBuild = path
			}
		}
	}
	if dataBuild == "" {
		if match := htmlDataBuildRE.FindStringSubmatch(html); len(match) >= 2 {
			dataBuild = strings.TrimSpace(match[1])
		}
	}
	if len(sources) == 0 {
		sources = []string{defaultPOWScript}
	}
	return sources, dataBuild
}

func timeMillis() int64 {
	return time.Now().UnixMilli()
}
