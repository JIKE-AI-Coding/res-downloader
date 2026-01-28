# Plugin 开发详解

> 本文档深入讲解 res-downloader 插件系统的开发方法,包含完整示例和最佳实践。

---

## 目录

- [插件系统架构](#插件系统架构)
- [Plugin 接口详解](#plugin-接口详解)
- [Bridge 桥接器](#bridge-桥接器)
- [完整开发示例](#完整开发示例)
- [高级技巧](#高级技巧)
- [调试方法](#调试方法)

---

## 插件系统架构

### 设计模式

插件系统采用了 **策略模式 (Strategy Pattern)** + **依赖注入 (Dependency Injection)** 的设计:

```
┌─────────────────────────────────────────────────────────┐
│                    Proxy Server                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Plugin Registry                     │    │
│  │  "qq.com"    → QqPlugin                         │    │
│  │  "douyin.com" → DouyinPlugin (假设)              │    │
│  │  "default"   → DefaultPlugin                    │    │
│  └─────────────────────────────────────────────────┘    │
│                          │                                │
│                          ▼                                │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Plugin Interface                    │    │
│  │  + Domains() []string                            │    │
│  │  + OnRequest(req, ctx) (req, resp)              │    │
│  │  + OnResponse(resp, ctx) resp                   │    │
│  └─────────────────────────────────────────────────┘    │
│                          ▲                                │
│                          │                                │
│  ┌───────────────────────┴───────────────────────┐      │
│  │                  Bridge                        │      │
│  │  注入核心能力,解耦插件与系统                    │      │
│  └───────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────┘
```

### 请求处理流程

```
HTTP Request → Proxy Server
    ↓
根据 Host 匹配 Plugin
    ↓
Plugin.OnRequest()
    ↓ (如果未返回 Response)
转发到目标服务器
    ↓
HTTP Response ← Proxy Server
    ↓
Plugin.OnResponse()
    ↓
如果返回 Response → 捕获资源
如果返回 nil → 继续传递
```

---

## Plugin 接口详解

### 接口定义

```go
// core/shared/plugin.go
type Plugin interface {
    // 1. 设置桥接器(注入依赖)
    SetBridge(*Bridge)

    // 2. 声明处理的域名
    // 返回值: ["qq.com", "douyin.com"]
    Domains() []string

    // 3. 请求拦截
    // 用途: 修改请求、拒绝请求、返回自定义响应
    // 返回值:
    //   (newReq, nil) - 修改后的请求,继续转发
    //   (nil, newResp) - 直接返回响应,不转发
    //   (nil, nil) - 不处理,使用默认逻辑
    OnRequest(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)

    // 4. 响应拦截
    // 用途: 提取资源、修改响应内容
    // 返回值:
    //   newResp - 修改后的响应
    //   nil - 不处理
    OnResponse(*http.Response, *goproxy.ProxyCtx) *http.Response
}
```

### 生命周期

```go
// 1. 插件注册(应用启动时)
func init() {
    ps := []shared.Plugin{
        &plugins.YourPlugin{},
    }
    // ...
}

// 2. 依赖注入
plugin.SetBridge(bridge)  // 在注册后自动调用

// 3. 域名匹配(每次请求时)
if plugin := matchPlugin(host); plugin != nil {
    // 4. 调用插件方法
    plugin.OnRequest(req, ctx)
    plugin.OnResponse(resp, ctx)
}
```

---

## Bridge 桥接器

### 设计目的

Bridge 将插件与核心系统解耦,使插件无需直接访问全局变量,便于测试和维护。

### 可用方法

```go
type Bridge struct {
    // 1. 版本信息
    GetVersion func() string

    // 2. 资源类型检查
    // 参数: "all", "video", "audio", "image" 等
    // 返回: (是否启用, 是否存在该类型)
    GetResType func(key string) (bool, bool)

    // 3. MIME 类型解析
    // 参数: "video/mp4"
    // 返回: ("video", ".mp4")
    TypeSuffix func(mime string) (string, string)

    // 4. 媒体标记(防重复)
    MediaIsMarked func(key string) bool
    MarkMedia     func(key string)

    // 5. 配置读取
    // 参数: "SaveDirectory", "TaskNumber" 等
    GetConfig func(key string) interface{}

    // 6. 向前端发送事件
    // 类型: "newResources", "downloadProgress"
    Send func(type string, data interface{})
}
```

### 使用示例

```go
func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 1. 检查资源类型是否启用
    isAll, _ := p.bridge.GetResType("all")
    isVideo, _ := p.bridge.GetResType("video")

    if !isAll && !isVideo {
        return nil  // 用户未勾选视频类型
    }

    // 2. 解析 MIME 类型
    classify, suffix := p.bridge.TypeSuffix(resp.Header.Get("Content-Type"))

    // 3. 防止重复
    urlSign := shared.Md5(resp.Request.URL.String())
    if p.bridge.MediaIsMarked(urlSign) {
        return nil  // 已处理过
    }

    // 4. 读取配置
    taskNumber := p.bridge.GetConfig("TaskNumber").(int)

    // 5. 发送资源到前端
    p.bridge.Send("newResources", shared.MediaInfo{...})

    // 6. 标记
    p.bridge.MarkMedia(urlSign)

    return nil
}
```

---

## 完整开发示例

### 场景: 支持 Bilibili 视频下载

#### 步骤 1: 创建插件文件

```go
// core/plugins/plugin.bilibili.com.go
package plugins

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "res-downloader/core/shared"
    "strings"

    gonanoid "github.com/matoous/go-nanoid/v2"
    "github.com/elazarl/goproxy"
)

// BilibiliPlugin 处理 Bilibili 域名
type BilibiliPlugin struct {
    bridge *shared.Bridge
}

// SetBridge 注入依赖
func (p *BilibiliPlugin) SetBridge(bridge *shared.Bridge) {
    p.bridge = bridge
}

// Domains 声明处理的域名
func (p *BilibiliPlugin) Domains() []string {
    return []string{"bilibili.com", "biliapi.com"}
}

// OnRequest 请求拦截
func (p *BilibiliPlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    // 场景 1: 拦截 API 请求,提取视频信息
    if strings.Contains(r.URL.Path, "/x/player/playurl") {
        // 读取请求体
        body, err := io.ReadAll(r.Body)
        if err != nil {
            return r, nil
        }

        // 恢复请求体(供后续使用)
        r.Body = io.NopCloser(strings.NewReader(string(body)))

        // 异步处理
        go p.handlePlayUrl(body, r)

        return r, nil
    }

    // 场景 2: 修改请求头(绕过限制)
    if strings.HasSuffix(r.Host, "biliapi.com") {
        r.Header.Set("Referer", "https://www.bilibili.com")
        r.Header.Set("Origin", "https://www.bilibili.com")
    }

    return nil, nil
}

// OnResponse 响应拦截
func (p *BilibiliPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    if resp.StatusCode != 200 {
        return nil
    }

    host := resp.Request.Host

    // 场景 1: 拦截视频流
    if strings.HasSuffix(host, "bilivideo.com") {
        contentType := resp.Header.Get("Content-Type")
        classify, suffix := p.bridge.TypeSuffix(contentType)

        if classify == "video" || classify == "audio" {
            // 检查是否启用
            isAll, _ := p.bridge.GetResType("all")
            isVideo, _ := p.bridge.GetResType("video")
            if !isAll && !isVideo {
                return nil
            }

            url := resp.Request.URL.String()
            urlSign := shared.Md5(url)

            // 防止重复
            if p.bridge.MediaIsMarked(urlSign) {
                return nil
            }

            // 生成 ID
            id, _ := gonanoid.New()

            // 构建资源信息
            mediaInfo := shared.MediaInfo{
                Id:          id,
                Url:         url,
                UrlSign:     urlSign,
                Domain:      shared.GetTopLevelDomain(url),
                Classify:    classify,
                Suffix:      suffix,
                ContentType: contentType,
                Status:      shared.DownloadStatusReady,
                OtherData:   make(map[string]string),
            }

            // 从 URL 中提取视频信息
            // 例如: https://xxx.bilivideo.com/xxx/video.m4s?range=...
            mediaInfo.Description = p.extractDescription(resp)

            // 发送到前端
            p.bridge.MarkMedia(urlSign)
            p.bridge.Send("newResources", mediaInfo)

            // 返回响应表示捕获
            return resp
        }
    }

    return nil
}

// handlePlayUrl 处理播放 URL API
func (p *BilibiliPlugin) handlePlayUrl(body []byte, r *http.Request) {
    // 解析 API 请求,获取视频 ID
    // 实际实现需要根据 Bilibili API 文档处理
}

// extractDescription 从响应中提取描述信息
func (p *BilibiliPlugin) extractDescription(resp *http.Response) string {
    // 可以从 Referer、Cookie 或之前的 API 响应中获取
    referer := resp.Request.Header.Get("Referer")

    // 提取视频 ID (例如: /video/BV1xx411c7mD)
    re := regexp.MustCompile(`/video/(BV\w+)`)
    matches := re.FindStringSubmatch(referer)
    if len(matches) > 1 {
        return fmt.Sprintf("Bilibili-%s", matches[1])
    }

    return "Bilibili Video"
}
```

#### 步骤 2: 注册插件

```go
// core/proxy.go
func init() {
    ps := []shared.Plugin{
        &plugins.QqPlugin{},
        &plugins.BilibiliPlugin{},  // 新增
        &plugins.DefaultPlugin{},
    }

    // ... 原有代码
}
```

#### 步骤 3: 测试验证

```bash
# 1. 重新编译
wails build

# 2. 启动应用
./res-downloader

# 3. 开启代理
点击"启动代理"按钮

# 4. 访问 Bilibili
打开浏览器,播放任意视频

# 5. 检查资源
查看应用内是否出现视频资源
```

---

## 高级技巧

### 1. 修改响应内容

**用途**: 注入 JS、修改 HTML 页面

```go
func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 只处理 HTML
    if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
        return nil
    }

    // 读取响应体
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil
    }

    // 修改内容
    newBody := strings.ReplaceAll(string(body),
        "</head>",
        "<script>console.log('Injected by plugin')</script></head>")

    // 构建新响应
    newResp := &http.Response{
        Status:     resp.Status,
        StatusCode: resp.StatusCode,
        Header:     resp.Header,
        Body:       io.NopCloser(strings.NewReader(newBody)),
    }

    // 更新 Content-Length
    newResp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

    return newResp
}
```

### 2. 并发安全

**问题**: 插件方法可能被并发调用

**解决**: 使用 sync.Map 保护共享状态

```go
type YourPlugin struct {
    bridge    *shared.Bridge
    cache     sync.Map  // 并发安全的缓存
}

func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    url := resp.Request.URL.String()

    // 使用 LoadOrStore 避免重复处理
    if _, loaded := p.cache.LoadOrStore(url, true); loaded {
        return nil  // 已处理过
    }

    // ... 业务逻辑

    return nil
}
```

### 3. 动态配置

**用途**: 根据用户配置调整插件行为

```go
func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 读取配置
    quality := p.bridge.GetConfig("Quality").(int)

    // 根据配置处理
    switch quality {
    case 0:  // 低清
        // ...
    case 1:  // 高清
        // ...
    }

    return nil
}
```

### 4. 错误处理

```go
func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    defer func() {
        if r := recover(); r != nil {
            globalLogger.Err().Msgf("Plugin panic: %v", r)
        }
    }()

    // 业务逻辑
    if err := someOperation(); err != nil {
        globalLogger.Err().Msgf("Plugin error: %v", err)
        return nil  // 不中断流程
    }

    return nil
}
```

---

## 调试方法

### 1. 添加日志

```go
import "res-downloader/core"

func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    core.globalLogger.Info().Msgf("Plugin processing: %s", resp.Request.URL.String())

    // 业务逻辑

    return nil
}
```

### 2. 导出网络请求

```bash
# 使用 curl 重现请求
# 从插件中导出完整请求信息
func (p *YourPlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    fmt.Printf("URL: %s\n", r.URL.String())
    fmt.Printf("Headers: %v\n", r.Header)
    fmt.Printf("Body: %v\n", r.Body)

    return nil, nil
}
```

### 3. 单元测试

```go
// core/plugins/plugin.bilibili.com_test.go
package plugins

import (
    "net/http"
    "net/url"
    "testing"
    "res-downloader/core/shared"
    "github.com/elazarl/goproxy"
)

func TestBilibiliPlugin_OnResponse(t *testing.T) {
    plugin := &BilibiliPlugin{}
    bridge := &shared.Bridge{
        GetResType: func(key string) (bool, bool) {
            return true, true
        },
        TypeSuffix: func(mime string) (string, string) {
            return "video", ".mp4"
        },
        MediaIsMarked: func(key string) bool {
            return false
        },
        MarkMedia: func(key string) {},
        Send: func(t string, d interface{}) {
            t.Logf("Sent event: %s, data: %v", t, d)
        },
    }
    plugin.SetBridge(bridge)

    // 构造测试请求
    req := &http.Request{
        URL: &url.URL{
            Host: "xxx.bilivideo.com",
            Path: "/video.mp4",
        },
    }

    resp := &http.Response{
        StatusCode: 200,
        Header: http.Header{
            "Content-Type": []string{"video/mp4"},
        },
        Request: req,
    }

    ctx := &goproxy.ProxyCtx{}

    // 调用插件
    result := plugin.OnResponse(resp, ctx)

    // 验证结果
    if result == nil {
        t.Error("Expected response, got nil")
    }
}
```

### 4. 抓包对比

```bash
# 使用 tcpdump 对比插件处理前后的请求
sudo tcpdump -i any -w before.pcap host bilibili.com

# 启用插件后再抓一次
sudo tcpdump -i any -w after.pcap host bilibili.com

# 使用 Wireshark 分析差异
```

---

## 最佳实践

### ✅ 推荐做法

1. **轻量级处理**: 插件只做资源提取,不做下载
2. **及时返回**: 处理完成后立即返回 nil,避免影响性能
3. **幂等性**: 多次调用同一请求应产生相同结果
4. **错误隔离**: 插件错误不应影响代理服务器
5. **配置驱动**: 通过配置控制行为,避免硬编码

### ❌ 避免做法

1. **阻塞操作**: 不要在插件中进行网络请求、文件 I/O
2. **全局状态**: 避免修改全局变量
3. **重复处理**: 使用 MarkMedia 防止重复
4. **忽略 Content-Type**: 根据类型判断是否处理
5. **硬编码域名**: 使用 Domains() 方法声明

---

## 常见场景代码片段

### 场景 1: 拦截 JSON API

```go
func (p *YourPlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    if strings.Contains(r.URL.Path, "/api/video") {
        body, _ := io.ReadAll(r.Body)
        var data map[string]interface{}
        json.Unmarshal(body, &data)

        // 提取视频 ID
        videoID := data["id"].(string)

        // 异步处理
        go p.fetchVideoInfo(videoID)

        // 恢复请求体
        r.Body = io.NopCloser(bytes.NewBuffer(body))
    }

    return nil, nil
}
```

### 场景 2: 修改 User-Agent

```go
func (p *YourPlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    if strings.HasSuffix(r.Host, "example.com") {
        r.Header.Set("User-Agent", "Mozilla/5.0 (Custom)")
    }
    return nil, nil
}
```

### 场景 3: 提取封面图

```go
func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 从 HTML 中提取封面
    if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
        body, _ := io.ReadAll(resp.Body)
        re := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
        matches := re.FindStringSubmatch(string(body))
        if len(matches) > 1 {
            coverURL := matches[1]
            // 保存到上下文或缓存
        }
    }

    return nil
}
```

---

**最后更新**: 2026-01-28
