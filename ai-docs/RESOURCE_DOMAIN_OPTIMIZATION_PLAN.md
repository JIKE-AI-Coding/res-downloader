# 资源列表域名中文展示优化方案

## 一、需求分析

### 1.1 当前问题
- 资源列表中域名展示格式为 `douyinvod.com:443`，不够直观
- 用户难以通过域名快速识别资源来源平台
- 需要将技术域名转换为用户友好的中文名称

### 1.2 改造目标
1. 将域名转换为中文平台名称展示（如：`douyinvod.com:443` → `抖音`）
2. 支持所有主流平台：微信视频号、小程序、抖音、快手、小红书、酷狗音乐、QQ音乐
3. 转换后仍支持按关键字检索功能（中文和域名都能搜索）
4. 添加完整的测试用例

---

## 二、现状分析

### 2.1 核心数据结构
**位置**：`core/shared/base.go:3-18`

```go
type MediaInfo struct {
    Id          string
    Url         string
    UrlSign     string
    CoverUrl    string
    Size        float64
    Domain      string  // 当前存储的是原始域名（如 douyinvod.com）
    Classify    string
    Suffix      string
    SavePath    string
    Status      string
    DecodeKey   string
    Description string
    ContentType string
    OtherData   map[string]string
}
```

### 2.2 域名提取逻辑
**位置**：`core/shared/utils.go:37-47`

```go
func GetTopLevelDomain(rawURL string) string {
    u, err := url.Parse(rawURL)
    if err == nil && u.Host != "" {
        rawURL = u.Host
    }
    domain, err := publicsuffix.EffectiveTLDPlusOne(rawURL)
    if err != nil {
        return rawURL
    }
    return domain
}
```

**特点**：
- 使用 `publicsuffix.EffectiveTLDPlusOne()` 提取有效顶级域名
- 例如：`https://v.douyin.com/video` → `douyin.com`
- 例如：`https://txmov2.a.kwimgs.com/` → `kwimgs.com`

### 2.3 当前插件注册的域名

| 插件 | 注册域名 | 实际处理范围 |
|------|----------|--------------|
| QqPlugin | qq.com | qq.com, channels.weixin.qq.com, finder.video.qq.com, res.wx.qq.com, wxapp.tc.qq.com |
| DefaultPlugin | default | 所有其他域名 |

### 2.4 前端展示逻辑
**位置**：`frontend/src/views/index.vue:243-291`

- 域名列直接绑定 `MediaInfo.Domain` 字段
- 支持悬停显示完整URL
- 当前搜索功能只搜索 `Url` 字段，不搜索 `Domain` 字段

### 2.5 搜索功能实现
**位置**：`frontend/src/views/index.vue:181-197`

当前搜索逻辑：
```javascript
if (urlSearchValue.value) {
    result = result.filter(item =>
        item.Url?.toLowerCase().includes(urlSearchValue.value.toLowerCase())
    )
}
```

**问题**：
- 只搜索 URL，不搜索 Domain
- 不支持中文平台名称搜索

---

## 三、改造方案设计

### 3.1 整体架构

采用 **数据层映射 + 展示层转换** 的双层架构：

```
┌─────────────────────────────────────────────────────────┐
│                    数据层 (Go Backend)                   │
│  MediaInfo.Domain 保持原始域名（如 douyin.com）         │
│  新增 MediaInfo.PlatformName 存储中文平台名（如 抖音）   │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  映射层 (Domain Mapping)                 │
│  DomainToPlatformName(domain string) → string           │
│  域名到中文名称的映射逻辑                                │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   展示层 (Vue Frontend)                  │
│  优先显示 PlatformName，回退显示 Domain                  │
│  搜索支持：中文平台名 + 原始域名 + URL                    │
└─────────────────────────────────────────────────────────┘
```

### 3.2 数据结构扩展

#### 3.2.1 MediaInfo 结构体扩展
**文件**：`core/shared/base.go`

```go
type MediaInfo struct {
    Id          string
    Url         string
    UrlSign     string
    CoverUrl    string
    Size        float64
    Domain      string        // 保留：原始域名（如 douyin.com）
    PlatformName string        // 新增：中文平台名（如 抖音）
    Classify    string
    Suffix      string
    SavePath    string
    Status      string
    DecodeKey   string
    Description string
    ContentType string
    OtherData   map[string]string
}
```

**设计说明**：
- `Domain` 字段保持不变，确保向后兼容
- `PlatformName` 为新增字段，存储中文平台名称
- JSON 序列化时自动包含新字段，前端可直接使用

### 3.3 域名映射逻辑

#### 3.3.1 平台域名映射表
**新建文件**：`core/shared/platform_mapping.go`

```go
package shared

// PlatformDomainMapping 定义域名到中文平台名称的映射
// 支持精确匹配和后缀匹配
var PlatformDomainMapping = map[string]string{
    // ===== 微信生态 =====
    "qq.com":                    "微信视频号",
    "weixin.qq.com":             "微信小程序",
    "mp.weixin.qq.com":          "微信公众号",
    "channels.weixin.qq.com":    "微信视频号",
    "finder.video.qq.com":       "微信视频号",
    "res.wx.qq.com":             "微信资源",
    "wxapp.tc.qq.com":           "微信小程序",

    // ===== 抖音 =====
    "douyin.com":                "抖音",
    "douyinvod.com":             "抖音",
    "aweme.snssdk.com":          "抖音",
    "douyinstatic.com":          "抖音",
    "douyinpic.com":             "抖音",
    "douyincdn.com":             "抖音",

    // ===== 快手 =====
    "kuaishou.com":              "快手",
    "kwimgs.com":                "快手",
    "ksycdn.com":                "快手",
    "kuaishouzt.com":            "快手",

    // ===== 小红书 =====
    "xiaohongshu.com":           "小红书",
    "xhslink.com":               "小红书",
    "xiaohongshu.com.cn":        "小红书",
    "edith.xiaohongshu.com":     "小红书",
    "sns-img-bd.xhscdn.com":     "小红书",
    "sns-video-bd.xhscdn.com":   "小红书",

    // ===== 酷狗音乐 =====
    "kugou.com":                 "酷狗音乐",
    "trackercdn.kugou.com":      "酷狗音乐",
    "mdn.kugou.com":             "酷狗音乐",

    // ===== QQ音乐 =====
    "qq.com":                    "QQ音乐",
    "y.qq.com":                  "QQ音乐",
    "qqmusic.qq.com":            "QQ音乐",
    "music.qq.com":              "QQ音乐",
    "streamoc.music.tc.qq.com":  "QQ音乐",
    "dl.stream.qqmusic.qq.com":  "QQ音乐",
    "isure.stream.qqmusic.qq.com": "QQ音乐",

    // ===== 其他平台 =====
    "bilibili.com":              "B站",
    "weibo.com":                 "微博",
    "zhihu.com":                 "知乎",
}

// GetPlatformName 根据域名获取中文平台名称
// 支持顶级域名匹配和子域名后缀匹配
func GetPlatformName(domain string) string {
    // 1. 精确匹配
    if name, ok := PlatformDomainMapping[domain]; ok {
        return name
    }

    // 2. 后缀匹配（处理子域名情况）
    for mappedDomain, platformName := range PlatformDomainMapping {
        if domain == mappedDomain ||
           strings.HasSuffix(domain, "."+mappedDomain) {
            return platformName
        }
    }

    // 3. 未匹配时返回原始域名
    return domain
}

// GetPlatformNameFromURL 从完整URL提取域名并转换为平台名称
func GetPlatformNameFromURL(rawURL string) string {
    domain := GetTopLevelDomain(rawURL)
    return GetPlatformName(domain)
}
```

#### 3.3.2 插件集成

**修改文件**：
1. `core/plugins/plugin.default.go:64`
2. `core/plugins/plugin.qq.com.go:170`

```go
// DefaultPlugin 示例
res := shared.MediaInfo{
    // ... 其他字段
    Domain:       shared.GetTopLevelDomain(rawUrl),
    PlatformName: shared.GetPlatformNameFromURL(rawUrl),  // 新增
    // ... 其他字段
}
```

### 3.4 前端展示改造

#### 3.4.1 类型定义更新
**文件**：`frontend/wailsjs/go/core/models.ts` 或前端自定义类型

```typescript
export interface MediaInfo {
    id: string;
    url: string;
    urlSign: string;
    coverUrl: string;
    size: number;
    domain: string;          // 原始域名
    platformName?: string;   // 中文平台名（新增）
    classify: string;
    suffix: string;
    savePath: string;
    status: string;
    decodeKey: string;
    description: string;
    contentType: string;
    otherData: Record<string, string>;
}
```

#### 3.4.2 域名列展示逻辑
**文件**：`frontend/src/views/index.vue`

```vue
{
  title: () => {
    if (checkedRowKeysValue.value.length > 0) {
      return h(NGradientText, {type: "success"}, t("index.choice") + `(${checkedRowKeysValue.value.length})`)
    }
    return h('div', {class: 'flex items-center'}, [
      t('index.domain'),
      h(NTooltip, {
        trigger: 'hover',
        placement: 'top'
      }, {
        trigger: () => h(NIcon, {size: 16, class: 'ml-2'}, () => h(SearchIcon)),
        default: () => t('index.search_domain_tooltip')
      })
    ])
  },
  key: "Domain",
  width: 120,
  render: (row: appType.MediaInfo) => {
    // 优先显示平台名称，回退到域名
    const displayText = row.PlatformName || row.Domain

    return h(NTooltip, {
      trigger: 'hover',
      placement: 'top'
    }, {
      trigger: () => h('span', {
        class: 'cursor-default'
      }, displayText),
      default: () => {
        // 悬停显示详细信息
        return h('div', {class: 'text-sm'}, [
          h('div', {}, `平台名称：${row.PlatformName || '未知'}`),
          h('div', {}, `原始域名：${row.Domain}`),
          h('div', {class: 'mt-1 text-xs opacity-70'}, row.Url)
        ])
      }
    })
  }
}
```

#### 3.4.3 搜索功能增强

**修改搜索逻辑**：
```javascript
const filteredData = computed(() => {
  let result = data.value

  // 1. 分类过滤
  if (filterClassify.value.length > 0) {
    result = result.filter(item => filterClassify.value.includes(item.Classify))
  }

  // 2. 描述搜索
  if (descriptionSearchValue.value) {
    result = result.filter(item =>
      item.Description?.toLowerCase().includes(descriptionSearchValue.value.toLowerCase())
    )
  }

  // 3. 域名/平台搜索（增强）
  if (urlSearchValue.value) {
    const searchLower = urlSearchValue.value.toLowerCase()
    result = result.filter(item => {
      // 支持中文平台名搜索
      if (item.PlatformName?.toLowerCase().includes(searchLower)) {
        return true
      }
      // 支持原始域名搜索
      if (item.Domain?.toLowerCase().includes(searchLower)) {
        return true
      }
      // 支持URL搜索
      if (item.Url?.toLowerCase().includes(searchLower)) {
        return true
      }
      return false
    })
  }

  return result
})
```

### 3.5 国际化支持

**文件**：`frontend/src/locales/zh.json`

```json
{
  "index": {
    "domain": "平台",
    "search_domain_tooltip": "支持搜索：平台名称、域名、URL"
  }
}
```

**文件**：`frontend/src/locales/en.json`

```json
{
  "index": {
    "domain": "Platform",
    "search_domain_tooltip": "Search by: platform name, domain, or URL"
  }
}
```

---

## 四、平台域名映射表

### 4.1 完整域名映射表

| 平台中文名 | 主要域名 | CDN/资源域名 | 匹配优先级 |
|-----------|---------|-------------|-----------|
| **微信视频号** | channels.weixin.qq.com<br/>finder.video.qq.com | res.wx.qq.com | 1 |
| **微信小程序** | mp.weixin.qq.com | wxapp.tc.qq.com | 1 |
| **微信公众号** | mp.weixin.qq.com | - | 1 |
| **微信资源** | - | res.wx.qq.com | 2 |
| **抖音** | douyin.com | douyinvod.com<br/>aweme.snssdk.com<br/>douyinstatic.com<br/>douyinpic.com<br/>douyincdn.com | 1 |
| **快手** | kuaishou.com | kwimgs.com<br/>ksycdn.com<br/>kuaishouzt.com | 1 |
| **小红书** | xiaohongshu.com | xhslink.com<br/>sns-img-bd.xhscdn.com<br/>sns-video-bd.xhscdn.com | 1 |
| **酷狗音乐** | kugou.com | trackercdn.kugou.com<br/>mdn.kugou.com | 1 |
| **QQ音乐** | y.qq.com | qqmusic.qq.com<br/>streamoc.music.tc.qq.com<br/>isure.stream.qqmusic.qq.com | 2（qq.com冲突，需特殊处理） |

### 4.2 特殊处理说明

#### 4.2.1 QQ 域名冲突问题
- `qq.com` 同时用于：QQ音乐、微信视频号、QQ微视
- **解决方案**：通过子域名优先级匹配
  - `y.qq.com` → QQ音乐
  - `channels.weixin.qq.com` → 微信视频号
  - `finder.video.qq.com` → 微信视频号
  - 其他 `*.qq.com` → QQ音乐（默认）

#### 4.2.2 域名匹配优先级
```
1. 完全匹配（精确）
   例如：channels.weixin.qq.com → 微信视频号

2. 后缀匹配（子域名）
   例如：v26-douyin-ving.cfcdn.com → 抖音（匹配 douyin.com 后缀）

3. 回退到原始域名
   例如：unknown-site.com → unknown-site.com
```

---

## 五、测试用例设计

### 5.1 单元测试（Go 后端）

#### 5.1.1 域名映射测试
**新建文件**：`core/shared/platform_mapping_test.go`

```go
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
        // ===== 微信生态 =====
        {"微信视频号 - channels", "channels.weixin.qq.com", "微信视频号"},
        {"微信视频号 - finder", "finder.video.qq.com", "微信视频号"},
        {"微信小程序", "mp.weixin.qq.com", "微信小程序"},
        {"微信资源", "res.wx.qq.com", "微信资源"},

        // ===== 抖音 =====
        {"抖音主域名", "douyin.com", "抖音"},
        {"抖音CDN - vod", "douyinvod.com", "抖音"},
        {"抖音SDK", "aweme.snssdk.com", "抖音"},

        // ===== 快手 =====
        {"快手主域名", "kuaishou.com", "快手"},
        {"快手CDN", "kwimgs.com", "快手"},

        // ===== 小红书 =====
        {"小红书主域名", "xiaohongshu.com", "小红书"},
        {"小红书CDN", "xhslink.com", "小红书"},

        // ===== 酷狗音乐 =====
        {"酷狗音乐", "kugou.com", "酷狗音乐"},

        // ===== QQ音乐 =====
        {"QQ音乐", "y.qq.com", "QQ音乐"},

        // ===== 未知域名 =====
        {"未知域名", "unknown-site.com", "unknown-site.com"},
        {"空字符串", "", ""},
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
        {"完整URL - 抖音", "https://v.douyin.com/video/123", "抖音"},
        {"完整URL - 快手", "https://live.kuaishou.com/live/123", "快手"},
        {"带端口的URL", "https://douyinvod.com:443/video", "抖音"},
        {"带路径的URL", "https://xiaohongshu.com/discovery/item/123", "小红书"},
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
        {"抖音子域名 - v26", "v26-douyin-ving.cfcdn.com", "抖音"},
        {"快手子域名 - stream", "stream.kwimgs.com", "快手"},
        {"小红书子域名 - sns", "sns-img-bd.xhscdn.com", "小红书"},
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
        {"QQ音乐主站", "y.qq.com", "QQ音乐"},
        {"微信视频号", "channels.weixin.qq.com", "微信视频号"},
        {"微信视频号 finder", "finder.video.qq.com", "微信视频号"},
        {"普通qq.com", "www.qq.com", "QQ音乐"},
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
```

#### 5.1.2 插件集成测试
**新建文件**：`core/plugins/plugin_mapping_test.go`

```go
package plugins

import (
    "res-downloader/core/shared"
    "testing"
)

func TestDefaultPluginPlatformMapping(t *testing.T) {
    plugin := &DefaultPlugin{}
    plugin.SetBridge(&shared.Bridge{})

    // 模拟测试 URL
    testURLs := []struct {
        url           string
        expectedDomain string
        expectedPlatform string
    }{
        {"https://v.douyin.com/123", "douyin.com", "抖音"},
        {"https://live.kuaishou.com/456", "kuaishou.com", "快手"},
        {"https://xiaohongshu.com/789", "xiaohongshu.com", "小红书"},
    }

    for _, tt := range testURLs {
        t.Run(tt.url, func(t *testing.T) {
            domain := shared.GetTopLevelDomain(tt.url)
            platform := shared.GetPlatformName(domain)

            if domain != tt.expectedDomain {
                t.Errorf("Domain mismatch: got %q, want %q", domain, tt.expectedDomain)
            }
            if platform != tt.expectedPlatform {
                t.Errorf("Platform mismatch: got %q, want %q", platform, tt.expectedPlatform)
            }
        })
    }
}
```

### 5.2 前端测试（Vue.js）

#### 5.2.1 组件测试
**新建文件**：`frontend/src/components/__tests__/DomainDisplay.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import IndexView from '@/views/index.vue'

describe('域名展示功能', () => {
  it('应该优先显示平台名称', () => {
    const mockData = {
      domain: 'douyin.com',
      platformName: '抖音',
      url: 'https://v.douyin.com/123'
    }

    const wrapper = mount(IndexView, {
      props: { data: [mockData] }
    })

    // 验证显示的是"抖音"而不是"douyin.com"
    expect(wrapper.text()).toContain('抖音')
  })

  it('平台名称为空时应显示域名', () => {
    const mockData = {
      domain: 'unknown-site.com',
      platformName: '',
      url: 'https://unknown-site.com/123'
    }

    const wrapper = mount(IndexView, {
      props: { data: [mockData] }
    })

    expect(wrapper.text()).toContain('unknown-site.com')
  })
})

describe('域名搜索功能', () => {
  it('应该支持中文平台名搜索', () => {
    const mockData = [
      { domain: 'douyin.com', platformName: '抖音', url: 'https://v.douyin.com/1' },
      { domain: 'kuaishou.com', platformName: '快手', url: 'https://live.kuaishou.com/2' }
    ]

    // 搜索"抖音"
    const result = mockData.filter(item =>
      item.platformName?.includes('抖音') ||
      item.domain?.includes('抖音') ||
      item.url?.includes('抖音')
    )

    expect(result).toHaveLength(1)
    expect(result[0].platformName).toBe('抖音')
  })

  it('应该支持原始域名搜索', () => {
    const mockData = [
      { domain: 'douyin.com', platformName: '抖音', url: 'https://v.douyin.com/1' }
    ]

    // 搜索"douyin"
    const result = mockData.filter(item =>
      item.platformName?.toLowerCase().includes('douyin') ||
      item.domain?.toLowerCase().includes('douyin') ||
      item.url?.toLowerCase().includes('douyin')
    )

    expect(result).toHaveLength(1)
  })
})
```

### 5.3 集成测试

#### 5.3.1 端到端测试场景

**测试场景 1：抖音资源检测和展示**
```
步骤：
1. 启动代理服务器
2. 访问抖音视频页面
3. 触发资源下载
4. 验证资源列表中显示"抖音"而非"douyinvod.com:443"
5. 验证搜索"抖音"能找到该资源
6. 验证搜索"douyin"也能找到该资源
```

**测试场景 2：未知域名处理**
```
步骤：
1. 访问未配置的网站资源
2. 验证显示原始域名
3. 验证搜索原始域名能找到资源
```

**测试场景 3：微信生态多平台识别**
```
步骤：
1. 访问微信视频号
2. 访问微信小程序
3. 验证两个资源分别显示"微信视频号"和"微信小程序"
4. 验证搜索"微信"能找到两个资源
```

---

## 六、实施步骤

### 阶段一：后端改造（2-3小时）

1. **创建平台映射模块**
   - 新建 `core/shared/platform_mapping.go`
   - 实现域名到中文名称的映射表
   - 实现 `GetPlatformName()` 和 `GetPlatformNameFromURL()` 函数

2. **扩展数据结构**
   - 修改 `core/shared/base.go`
   - 在 `MediaInfo` 中添加 `PlatformName` 字段

3. **集成到插件**
   - 修改 `core/plugins/plugin.default.go:64`
   - 修改 `core/plugins/plugin.qq.com.go:170`
   - 在创建 `MediaInfo` 时设置 `PlatformName` 字段

4. **编写单元测试**
   - 创建 `core/shared/platform_mapping_test.go`
   - 覆盖所有平台域名映射场景
   - 测试边界情况（空值、未知域名等）

5. **运行测试**
   ```bash
   cd /Users/wu/Desktop/aicoding/opensource/res-downloader
   go test ./core/shared/... -v
   go test ./core/plugins/... -v
   ```

### 阶段二：前端改造（1-2小时）

1. **更新类型定义**
   - 更新 `frontend/wailsjs/go/core/models.ts`（如果自动生成）
   - 或在前端创建自定义类型

2. **修改域名展示逻辑**
   - 修改 `frontend/src/views/index.vue:243-291`
   - 实现优先显示 `PlatformName`，回退显示 `Domain`
   - 增强悬停提示信息

3. **增强搜索功能**
   - 修改 `frontend/src/views/index.vue:181-197`
   - 支持搜索 `PlatformName`、`Domain`、`Url` 三个字段

4. **添加国际化**
   - 更新 `frontend/src/locales/zh.json`
   - 更新 `frontend/src/locales/en.json`

5. **前端测试**
   ```bash
   cd frontend
   npm run test
   ```

### 阶段三：联调测试（1小时）

1. **启动开发环境**
   ```bash
   wails dev
   ```

2. **端到端测试**
   - 测试各平台资源识别
   - 测试搜索功能
   - 测试边界情况

3. **问题修复**
   - 根据测试结果调整映射表
   - 优化搜索逻辑

### 阶段四：文档更新（30分钟）

1. **更新开发文档**
   - 在 `ai-docs/` 中记录本次改造
   - 更新插件开发指南（如有）

2. **更新用户文档**
   - 说明平台名称展示功能
   - 说明搜索功能增强

---

## 七、风险评估与应对

### 7.1 风险点

| 风险 | 影响 | 应对措施 |
|-----|------|---------|
| **域名映射表不完整** | 部分资源仍显示域名 | 使用原始域名作为回退，不影响功能 |
| **qq.com 域名冲突** | QQ音乐/微信识别错误 | 通过子域名优先级匹配解决 |
| **JSON 兼容性** | 旧版本客户端不识别新字段 | 新字段为可选，旧版本自动忽略 |
| **性能影响** | 域名匹配增加耗时 | 匹配逻辑简单（O(n)），影响可忽略 |
| **搜索性能** | 多字段搜索变慢 | 用户数据量小（<1000条），影响可忽略 |

### 7.2 回退方案

如果出现问题，可以通过以下方式回退：
1. **后端**：移除 `PlatformName` 赋值，字段为空时前端自动显示 `Domain`
2. **前端**：恢复原搜索逻辑（只搜索 URL）
3. **配置**：通过配置开关控制是否启用中文平台名功能

---

## 八、扩展性考虑

### 8.1 未来扩展

1. **配置化映射表**
   - 将域名映射表移到配置文件
   - 支持用户自定义平台名称

2. **图标支持**
   - 为每个平台添加图标
   - 在域名列显示平台图标

3. **多语言支持**
   - 将平台名称移到 i18n 系统
   - 支持英文、日文等多语言

4. **智能识别**
   - 基于 URL 模式识别平台
   - 支持正则表达式匹配

### 8.2 维护指南

**添加新平台步骤**：
1. 在 `PlatformDomainMapping` 中添加域名映射
2. 在测试文件中添加对应测试用例
3. 运行测试确保通过
4. 更新文档

**更新平台名称步骤**：
1. 修改 `PlatformDomainMapping` 中的值
2. 运行测试验证
3. 重新构建应用

---

## 九、关键文件清单

### 需要修改的文件

| 文件路径 | 修改内容 | 优先级 |
|---------|---------|--------|
| `core/shared/base.go` | MediaInfo 添加 PlatformName 字段 | 高 |
| `core/shared/platform_mapping.go` | **新建**：域名映射逻辑 | 高 |
| `core/shared/platform_mapping_test.go` | **新建**：单元测试 | 高 |
| `core/plugins/plugin.default.go` | 集成平台名称设置 | 高 |
| `core/plugins/plugin.qq.com.go` | 集成平台名称设置 | 高 |
| `frontend/src/views/index.vue` | 域名展示 + 搜索增强 | 高 |
| `frontend/src/locales/zh.json` | 国际化文本 | 中 |
| `frontend/src/locales/en.json` | 国际化文本 | 中 |

### 不需要修改的文件

- `core/shared/utils.go` - `GetTopLevelDomain()` 保持不变
- `core/proxy.go` - 插件注册逻辑不变
- `core/resource.go` - 资源管理逻辑不变

---

## 十、验收标准

### 10.1 功能验收

- [ ] 资源列表中域名列显示中文平台名称（如"抖音"）
- [ ] 未配置的平台显示原始域名
- [ ] 搜索"抖音"能找到抖音平台资源
- [ ] 搜索"douyin"也能找到抖音平台资源
- [ ] 悬停域名显示完整信息（平台名 + 域名 + URL）
- [ ] 所有7个主流平台都能正确识别

### 10.2 测试验收

- [ ] 单元测试覆盖率 > 90%
- [ ] 所有单元测试通过
- [ ] 端到端测试场景全部通过
- [ ] 无性能回归

### 10.3 兼容性验收

- [ ] 旧版本资源数据能正常显示（回退到域名）
- [ ] JSON 序列化/反序列化正常
- [ ] 前后端通信正常

---

## 十一、预期效果

### 改造前
```
┌──────────────────────┬──────────────┬────────┐
│ 平台                 │ 大小         │ 操作   │
├──────────────────────┼──────────────┼────────┤
│ douyinvod.com:443    │ 15.2MB       │ 下载   │
│ kwimgs.com           │ 8.5MB        │ 下载   │
│ xiaohongshu.com      │ 12.1MB       │ 下载   │
└──────────────────────┴──────────────┴────────┘
```

### 改造后
```
┌──────────────────────┬──────────────┬────────┐
│ 平台                 │ 大小         │ 操作   │
├──────────────────────┼──────────────┼────────┤
│ 抖音                 │ 15.2MB       │ 下载   │
│ 快手                 │ 8.5MB        │ 下载   │
│ 小红书               │ 12.1MB       │ 下载   │
└──────────────────────┴──────────────┴────────┘
```

### 搜索体验提升
- 搜索"抖音" → 找到所有抖音资源
- 搜索"douyin" → 找到所有抖音资源
- 搜索"微信" → 找到微信视频号、小程序、公众号所有资源

---

## 十二、总结

本方案通过 **后端数据层扩展 + 前端展示层优化** 的方式，实现了资源列表域名的中文展示，同时保持了向后兼容性和扩展性。

**核心优势**：
1. ✅ 用户体验友好：中文平台名称直观易懂
2. ✅ 搜索功能增强：支持中文和域名双模式搜索
3. ✅ 向后兼容：不影响现有功能和数据
4. ✅ 易于维护：映射表集中管理，添加新平台简单
5. ✅ 测试完善：单元测试 + 集成测试覆盖全面

**预计工作量**：4-6 小时
**风险等级**：低
**推荐优先级**：高
