package shared

import "strings"

// PlatformDomainMapping defines domain to Chinese platform name mapping
// Supports exact match and suffix match for subdomains
var PlatformDomainMapping = map[string]string{
	// ===== WeChat Ecosystem =====
	"weixin.qq.com":          "微信小程序",
	"mp.weixin.qq.com":       "微信公众号",
	"channels.weixin.qq.com": "微信视频号",
	"finder.video.qq.com":    "微信视频号",
	"res.wx.qq.com":          "微信资源",
	"wxapp.tc.qq.com":        "微信小程序",

	// ===== Douyin (TikTok China) =====
	"douyin.com":       "抖音",
	"douyinvod.com":    "抖音",
	"aweme.snssdk.com": "抖音",
	"douyinstatic.com": "抖音",
	"douyinpic.com":    "抖音",
	"douyincdn.com":    "抖音",

	// ===== Kuaishou =====
	"kuaishou.com":   "快手",
	"kwimgs.com":     "快手",
	"ksycdn.com":     "快手",
	"kuaishouzt.com": "快手",

	// ===== Xiaohongshu (Little Red Book) =====
	"xiaohongshu.com":         "小红书",
	"xhslink.com":             "小红书",
	"xiaohongshu.com.cn":      "小红书",
	"edith.xiaohongshu.com":   "小红书",
	"sns-img-bd.xhscdn.com":   "小红书",
	"sns-video-bd.xhscdn.com": "小红书",

	// ===== Kugou Music =====
	"kugou.com":            "酷狗音乐",
	"trackercdn.kugou.com": "酷狗音乐",
	"mdn.kugou.com":        "酷狗音乐",

	// ===== QQ Music =====
	"y.qq.com":                    "QQ音乐",
	"qqmusic.qq.com":              "QQ音乐",
	"music.qq.com":                "QQ音乐",
	"streamoc.music.tc.qq.com":    "QQ音乐",
	"dl.stream.qqmusic.qq.com":    "QQ音乐",
	"isure.stream.qqmusic.qq.com": "QQ音乐",

	// ===== Other Platforms =====
	"bilibili.com": "B站",
	"weibo.com":    "微博",
	"zhihu.com":    "知乎",
}

// GetPlatformName returns Chinese platform name for a given domain
// Supports top-level domain matching and subdomain suffix matching
func GetPlatformName(domain string) string {
	if domain == "" {
		return domain
	}

	// 1. Exact match
	if name, ok := PlatformDomainMapping[domain]; ok {
		return name
	}

	// 2. Suffix match (handle subdomain cases)
	for mappedDomain, platformName := range PlatformDomainMapping {
		// Check if it's an exact match or ends with .mappedDomain
		if domain == mappedDomain || suffixMatch(domain, mappedDomain) {
			return platformName
		}
	}

	// 3. Return original domain if no match found
	return domain
}

// suffixMatch checks if domain ends with .mappedDomain (for subdomain matching)
// Example: v26.douyin.com should match douyin.com
func suffixMatch(domain, mappedDomain string) bool {
	// More specific subdomain check - only match if mappedDomain is not a suffix of another entry
	// For example, we want "finder.video.qq.com" to match "qq.com" but with proper priority
	return len(domain) > len(mappedDomain)+1 &&
		domain[len(domain)-len(mappedDomain)-1:] == "."+mappedDomain
}

// GetPlatformNameFromURL extracts domain from URL and converts to platform name
func GetPlatformNameFromURL(rawURL string) string {
	domain := GetTopLevelDomain(rawURL)
	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	return GetPlatformName(domain)
}
