package shared

import (
	"testing"
)

func TestGetPlatformName(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		// ===== WeChat Ecosystem =====
		{"WeChat Channels - channels", "channels.weixin.qq.com", "微信视频号"},
		{"WeChat Channels - finder", "finder.video.qq.com", "微信视频号"},
		{"WeChat Official Account", "mp.weixin.qq.com", "微信公众号"},
		{"WeChat Resources", "res.wx.qq.com", "微信资源"},
		{"WeChat Mini App", "wxapp.tc.qq.com", "微信小程序"},
		{"Weixin domain", "weixin.qq.com", "微信小程序"},

		// ===== Douyin (TikTok China) =====
		{"Douyin main domain", "douyin.com", "抖音"},
		{"Douyin CDN - vod", "douyinvod.com", "抖音"},
		{"Douyin SDK", "aweme.snssdk.com", "抖音"},
		{"Douyin Static", "douyinstatic.com", "抖音"},
		{"Douyin Pic", "douyinpic.com", "抖音"},
		{"Douyin CDN", "douyincdn.com", "抖音"},

		// ===== Kuaishou =====
		{"Kuaishou main domain", "kuaishou.com", "快手"},
		{"Kuaishou CDN", "kwimgs.com", "快手"},
		{"Kuaishou CDN 2", "ksycdn.com", "快手"},
		{"Kuaishou ZT", "kuaishouzt.com", "快手"},

		// ===== Xiaohongshu (Little Red Book) =====
		{"Xiaohongshu main domain", "xiaohongshu.com", "小红书"},
		{"Xiaohongshu CDN", "xhslink.com", "小红书"},
		{"Xiaohongshu China", "xiaohongshu.com.cn", "小红书"},
		{"Xiaohongshu Edith", "edith.xiaohongshu.com", "小红书"},
		{"Xiaohongshu Image CDN", "sns-img-bd.xhscdn.com", "小红书"},
		{"Xiaohongshu Video CDN", "sns-video-bd.xhscdn.com", "小红书"},

		// ===== Kugou Music =====
		{"Kugou Music", "kugou.com", "酷狗音乐"},
		{"Kugou Tracker CDN", "trackercdn.kugou.com", "酷狗音乐"},
		{"Kugou MDN", "mdn.kugou.com", "酷狗音乐"},

		// ===== QQ Music =====
		{"QQ Music", "y.qq.com", "QQ音乐"},
		{"QQ Music Domain", "qqmusic.qq.com", "QQ音乐"},
		{"QQ Music 2", "music.qq.com", "QQ音乐"},
		{"QQ Music Stream", "streamoc.music.tc.qq.com", "QQ音乐"},
		{"QQ Music DL", "dl.stream.qqmusic.qq.com", "QQ音乐"},
		{"QQ Music ISure", "isure.stream.qqmusic.qq.com", "QQ音乐"},

		// ===== Other Platforms =====
		{"Bilibili", "bilibili.com", "B站"},
		{"Weibo", "weibo.com", "微博"},
		{"Zhihu", "zhihu.com", "知乎"},

		// ===== Unknown Domains =====
		{"Unknown domain", "unknown-site.com", "unknown-site.com"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlatformName(tt.domain)
			if result != tt.expected {
				t.Errorf("GetPlatformName(%q) = %q, want %q",
					tt.domain, result, tt.expected)
			}
		})
	}
}

func TestGetPlatformNameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"Full URL - Douyin", "https://v.douyin.com/video/123", "抖音"},
		{"Full URL - Kuaishou", "https://live.kuaishou.com/live/123", "快手"},
		{"URL with port", "https://douyinvod.com:443/video", "抖音"},
		{"URL with path", "https://xiaohongshu.com/discovery/item/123", "小红书"},
		{"WeChat Channels URL", "https://channels.weixin.qq.com/video/123", "qq.com"},
		{"QQ Music URL", "https://y.qq.com/n/ryqq/songDetail/123", "qq.com"},
		{"Kugou URL", "https://www.kugou.com/song/123", "酷狗音乐"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlatformNameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("GetPlatformNameFromURL(%q) = %q, want %q",
					tt.url, result, tt.expected)
			}
		})
	}
}

func TestSubdomainMatching(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{"Douyin subdomain - v26", "v26-douyin-ving.cfcdn.com", "v26-douyin-ving.cfcdn.com"}, // Won't match without explicit mapping
		{"Kuaishou subdomain - stream", "stream.kwimgs.com", "快手"},                       // Will match via suffix
		{"Xiaohongshu subdomain - sns", "sns-img-bd.xhscdn.com", "小红书"},                  // Explicit mapping
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlatformName(tt.domain)
			if result != tt.expected {
				t.Errorf("GetPlatformName(%q) = %q, want %q",
					tt.domain, result, tt.expected)
			}
		})
	}
}

func TestQQDomainConflict(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{"QQ Music main", "y.qq.com", "QQ音乐"},
		{"WeChat Channels", "channels.weixin.qq.com", "微信视频号"},
		{"WeChat Channels finder", "finder.video.qq.com", "微信视频号"},
		{"WeChat Official Account", "mp.weixin.qq.com", "微信公众号"},
		{"Generic qq.com", "www.qq.com", "www.qq.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlatformName(tt.domain)
			if result != tt.expected {
				t.Errorf("GetPlatformName(%q) = %q, want %q",
					tt.domain, result, tt.expected)
			}
		})
	}
}

func TestSuffixMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		domain       string
		mappedDomain string
		expected     bool
	}{
		{"Exact match", "douyin.com", "douyin.com", false},
		{"Valid subdomain", "v.douyin.com", "douyin.com", true},
		{"Deep subdomain", "a.b.c.douyin.com", "douyin.com", true},
		{"Different domain", "kuaishou.com", "douyin.com", false},
		{"Partial match", "mydouyin.com", "douyin.com", false},
		{"Empty mapped", "douyin.com", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suffixMatch(tt.domain, tt.mappedDomain)
			if result != tt.expected {
				t.Errorf("suffixMatch(%q, %q) = %v, want %v",
					tt.domain, tt.mappedDomain, result, tt.expected)
			}
		})
	}
}

func TestPlatformMappingCompleteness(t *testing.T) {
	// Test that all major platforms have at least one domain mapping
	requiredPlatforms := []struct {
		name      string
		domains   []string
		mustExist bool
	}{
		{"抖音", []string{"douyin.com", "douyinvod.com"}, true},
		{"快手", []string{"kuaishou.com", "kwimgs.com"}, true},
		{"小红书", []string{"xiaohongshu.com"}, true},
		{"酷狗音乐", []string{"kugou.com"}, true},
		{"微信视频号", []string{"channels.weixin.qq.com", "finder.video.qq.com"}, true},
		{"微信公众号", []string{"mp.weixin.qq.com"}, true},
		{"微信小程序", []string{"weixin.qq.com", "wxapp.tc.qq.com"}, true},
	}

	for _, platform := range requiredPlatforms {
		t.Run(platform.name, func(t *testing.T) {
			found := false
			for _, domain := range platform.domains {
				if name, ok := PlatformDomainMapping[domain]; ok {
					found = true
					if name != platform.name {
						t.Errorf("Domain %q mapped to %q, expected %q", domain, name, platform.name)
					}
				}
			}
			if platform.mustExist && !found {
				t.Errorf("No domain found for platform %q", platform.name)
			}
		})
	}
}
