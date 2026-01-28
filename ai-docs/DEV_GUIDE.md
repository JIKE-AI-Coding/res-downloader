# Res-Downloader 二次开发指南

> 本文档为 res-downloader 项目的详细技术分析报告,旨在帮助开发者快速理解项目架构并进行二次开发。

---

## 目录

- [项目整体概览](#项目整体概览)
- [核心技术栈](#核心技术栈)
- [架构设计](#架构设计)
- [模块说明](#模块说明)
- [开发环境搭建](#开发环境搭建)
- [扩展开发指南](#扩展开发指南)
- [常见问题](#常见问题)

---

## 项目整体概览

### 项目定位

**res-downloader** 是一款基于 Go + Wails 框架开发的跨平台资源下载工具,通过**HTTP/HTTPS 代理拦截**的方式实现网络资源嗅探与下载。

### 核心功能

1. **资源拦截**: 通过系统代理拦截网络流量,捕获视频、音频、图片等资源
2. **多格式支持**: 支持 mp4、m3u8、flv、mp3、图片等多种媒体格式
3. **平台适配**: 针对微信视频号、抖音、快手等平台做了特殊处理
4. **高速下载**: 支持多线程分片下载、断点续传
5. **视频解密**: 支持微信视频号加密视频的解密

### 技术亮点

- **插件化架构**: 通过 Plugin 接口实现对不同平台的定制化处理
- **MITM 代理**: 使用中间人攻击技术拦截 HTTPS 流量
- **跨平台 GUI**: Wails 框架实现桌面应用,前端使用 Vue 3
- **WebSocket 通信**: Go 后端与前端通过实时事件通信

---

## 核心技术栈

| 层级 | 技术栈 | 说明 |
|------|--------|------|
| **后端** | Go 1.x | 核心业务逻辑 |
| **桌面框架** | Wails v2 | Go + Web 技术构建跨平台桌面应用 |
| **代理服务** | goproxy | HTTP/HTTPS 代理服务器 |
| **前端** | Vue 3 + TypeScript | 用户界面 |
| **UI 组件** | Naive UI | Vue 3 组件库 |
| **状态管理** | Pinia | 前端状态管理 |
| **构建工具** | Vite | 前端构建工具 |
| **样式** | Tailwind CSS | 原子化 CSS 框架 |

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        桌面应用窗口                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   Vue 3 前端界面                      │  │
│  │  - 资源列表展示                                        │  │
│  │  - 设置管理                                           │  │
│  │  - 下载控制                                           │  │
│  └─────────────┬────────────────────────────────────────┘  │
│                │ Wails Bridge (JS ↔ Go)                    │
│  ┌─────────────▼────────────────────────────────────────┐  │
│  │                   Go 后端服务                         │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ HTTP Server  │  │   Proxy      │  │  Resource   │ │  │
│  │  │   (API)      │  │  (goproxy)   │  │  Manager    │ │  │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │  │
│  │  ┌─────────────────────────────────────────────────┐ │  │
│  │  │              Plugin System                      │ │  │
│  │  │  - QQ Plugin (微信视频号)                        │ │  │
│  │  │  - Default Plugin (通用处理)                     │ │  │
│  │  └─────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
                   ┌───────────────┐
                   │  系统代理设置  │
                   │  (HTTP/HTTPS) │
                   └───────────────┘
                           │
                           ▼
                   ┌───────────────────────────────┐
                   │      外部网络流量              │
                   │  (浏览器、应用程序的网络请求)   │
                   └───────────────────────────────┘
```

### 数据流向

```
用户访问网页
    ↓
系统代理 (127.0.0.1:8899)
    ↓
goproxy 代理服务器拦截
    ↓
Plugin 处理 (根据域名匹配)
    ↓
提取资源 URL → 通过 WebSocket 发送到前端
    ↓
前端展示资源列表
    ↓
用户点击下载
    ↓
Go 后端多线程下载 → 保存到本地
```

---

## 模块说明

### 目录结构

```
res-downloader/
├── main.go                 # 应用入口,Wails 配置
├── core/                   # Go 后端核心代码
│   ├── app.go             # 应用主结构,生命周期管理
│   ├── proxy.go           # 代理服务器核心逻辑
│   ├── http.go            # HTTP API 服务
│   ├── resource.go        # 资源管理,下载任务处理
│   ├── downloader.go      # 多线程下载器实现
│   ├── config.go          # 配置管理
│   ├── bind.go            # Wails 绑定接口
│   ├── storage.go         # 本地存储封装
│   ├── rule.go            # 域名规则管理
│   ├── system_*.go        # 系统代理设置(平台相关)
│   ├── plugins/           # 插件系统
│   │   ├── plugin.qq.com.go    # 微信视频号插件
│   │   └── plugin.default.go   # 默认插件
│   └── shared/            # 共享代码
│       ├── plugin.go      # Plugin 接口定义
│       ├── const.go       # 常量定义
│       └── utils.go       # 工具函数
├── frontend/              # Vue 3 前端代码
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   ├── components/    # 可复用组件
│   │   ├── stores/        # Pinia 状态管理
│   │   ├── api/           # API 封装
│   │   └── router/        # 路由配置
│   └── package.json
└── wails.json             # Wails 项目配置
```

### 核心模块详解

#### 1. 应用入口 (main.go)

**职责**: 初始化 Wails 应用,配置窗口、菜单、生命周期钩子

**关键代码**:
```go
// 创建应用实例
app := core.GetApp(assets, wailsJson)
bind := core.NewBind()

// 启动时执行
OnStartup: func(ctx context.Context) {
    app.Startup(ctx)  // 启动 HTTP 服务
}

// 退出时执行
OnShutdown: func(ctx context.Context) {
    app.OnExit()  // 清理资源,关闭代理
}
```

**二次开发要点**:
- 修改窗口大小: 调整 `Width`、`Height` 参数
- 添加菜单项: 在 `appMenu` 中追加
- 自定义图标: 替换 `build/appicon.png`

#### 2. 代理模块 (core/proxy.go)

**职责**: HTTP/HTTPS 代理服务器,流量拦截核心

**关键组件**:

| 组件 | 说明 |
|------|------|
| `Proxy` 结构体 | 代理服务器实例 |
| `pluginRegistry` | 域名 → Plugin 映射表 |
| `httpRequestEvent` | 请求拦截处理 |
| `httpResponseEvent` | 响应拦截处理 |

**核心流程**:

```go
// 1. 初始化代理
proxyOnce.Startup()

// 2. 设置 CA 证书(用于 HTTPS MITM)
p.setCa()

// 3. 配置上游代理(可选)
p.setTransport()

// 4. 注册 CONNECT 处理(决定是否 MITM)
p.Proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) {
    if ruleOnce.shouldMitm(host) {
        return goproxy.MitmConnect, host  // 中间人拦截
    }
    return goproxy.OkConnect, host        // 直接转发
})

// 5. 请求/响应处理
p.Proxy.OnRequest().DoFunc(p.httpRequestEvent)
p.Proxy.OnResponse().DoFunc(p.httpResponseEvent)
```

**二次开发要点**:
- **新增插件**: 在 `init()` 函数中注册到 `pluginRegistry`
- **修改拦截规则**: 调整 `rule.go` 中的规则逻辑
- **自定义 CA 证书**: 修改 `app.go` 中的 `PublicCrt` 和 `PrivateKey`

#### 3. 插件系统 (core/plugins/)

**设计模式**: 策略模式 + 桥接模式

**Plugin 接口**:

```go
// core/shared/plugin.go
type Plugin interface {
    SetBridge(*Bridge)                    // 注入依赖
    Domains() []string                    // 处理的域名列表
    OnRequest(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
    OnResponse(*http.Response, *goproxy.ProxyCtx) *http.Response
}
```

**Bridge 桥接器**: 提供插件所需的核心能力

| 方法 | 说明 |
|------|------|
| `GetVersion()` | 获取应用版本 |
| `GetResType(key)` | 检查资源类型是否启用 |
| `TypeSuffix(mime)` | 根据 MIME 类型获取后缀 |
| `MediaIsMarked(key)` | 检查资源是否已标记 |
| `MarkMedia(key)` | 标记资源(避免重复) |
| `GetConfig(key)` | 获取配置项 |
| `Send(type, data)` | 向前端发送事件 |

**示例: 微信视频号插件 (plugin.qq.com.go)**

```go
type QqPlugin struct {
    bridge *shared.Bridge
}

func (p *QqPlugin) Domains() []string {
    return []string{"qq.com"}
}

func (p *QqPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 1. 检查 Content-Type
    classify, _ := p.bridge.TypeSuffix(resp.Header.Get("Content-Type"))

    // 2. 嗅探视频资源
    if classify == "video" && strings.HasSuffix(resp.Request.Host, "finder.video.qq.com") {
        return resp  // 返回响应表示捕获该资源
    }

    // 3. 注入 JS (获取更多信息)
    if strings.HasSuffix(host, "channels.weixin.qq.com") {
        return p.replaceWxJsContent(resp, ".js\"", ".js?v="+p.v()+"\"")
    }

    return nil  // 不处理
}
```

**二次开发要点**:
- **新增平台支持**: 创建新插件文件,实现 `Plugin` 接口
- **修改抓取逻辑**: 在 `OnResponse` 中调整资源提取逻辑
- **自定义请求处理**: 在 `OnRequest` 中修改请求(如添加 Header)

#### 4. 下载器模块 (core/downloader.go)

**职责**: 多线程分片下载,支持断点续传

**核心特性**:

- **自动分片**: 根据文件大小自动计算分片数量
- **并发下载**: 使用 Goroutine 并发下载分片
- **错误重试**: 失败自动重试,最多 3 次
- **进度回调**: 实时上报下载进度
- **取消支持**: 支持中途取消下载

**关键结构**:

```go
type FileDownloader struct {
    Url              string              // 下载地址
    FileName         string              // 保存路径
    totalTasks       int                 // 分片数量
    IsMultiPart      bool                // 是否多线程
    TotalSize        int64               // 文件总大小
    DownloadTaskList []*DownloadTask     // 分片任务列表
    progressCallback ProgressCallback    // 进度回调
    ctx              context.Context     // 取消上下文
}
```

**下载流程**:

```
1. init() - HEAD 请求获取文件大小
   ↓
2. createDownloadTasks() - 创建分片任务
   ↓
3. startDownload() - 并发执行任务
   ↓
4. verifyDownload() - 验证完整性
```

**二次开发要点**:
- **调整线程数**: 修改 `config.go` 中的 `TaskNumber`
- **自定义重试策略**: 调整 `MaxRetries` 和 `RetryDelay`
- **添加限速**: 在 `doDownloadTask` 中添加速率控制

#### 5. HTTP API 服务 (core/http.go)

**职责**: 为前端提供 RESTful API,处理文件对话框等系统调用

**主要接口**:

| 路由 | 方法 | 说明 |
|------|------|------|
| `/api/install` | POST | 安装证书 |
| `/api/proxy/open` | POST | 开启系统代理 |
| `/api/proxy/close` | POST | 关闭系统代理 |
| `/api/config` | GET/POST | 获取/设置配置 |
| `/api/download` | POST | 开始下载 |
| `/api/cancel` | POST | 取消下载 |
| `/api/cert` | GET | 下载 CA 证书 |

**Wails 事件通信**:

```go
// 向前端发送事件
httpServerOnce.send("newResources", mediaInfo)
httpServerOnce.send("downloadProgress", progress)
```

**二次开发要点**:
- **新增 API**: 在 `http.go` 中添加处理函数
- **修改事件格式**: 调整 `send()` 调用的数据结构
- **自定义中间件**: 修改 `middleware.go`

#### 6. 前端模块 (frontend/)

**技术栈**: Vue 3 + TypeScript + Naive UI + Pinia

**核心文件**:

| 文件 | 说明 |
|------|------|
| `src/views/index.vue` | 主页面,资源列表与下载控制 |
| `src/views/setting.vue` | 设置页面 |
| `src/stores/event.ts` | 事件监听,处理后端推送 |
| `src/stores/index.ts` | 全局状态管理 |
| `src/api/app.ts` | API 请求封装 |
| `src/router/index.ts` | 路由配置 |

**事件驱动架构**:

```typescript
// stores/event.ts
const init = () => {
    EventsOn("event", (res: any) => {
        const data = JSON.parse(res)
        if (handles.value.hasOwnProperty(data.type)) {
            handles.value[data.type](data.data)  // 分发事件
        }
    })
}

// 注册事件处理器
eventStore.addHandle({
    type: "newResources",
    event: (res: appType.MediaInfo) => {
        data.value.push(res)
    }
})
```

**二次开发要点**:
- **新增页面**: 在 `views/` 下创建组件,在 `router/index.ts` 中注册
- **新增事件处理器**: 调用 `eventStore.addHandle()`
- **修改 UI**: 使用 Naive UI 组件,遵循现有样式规范

---

## 开发环境搭建

### 前置要求

- **Go**: 1.18 或更高版本
- **Node.js**: 16.x 或更高版本
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/putyy/res-downloader.git
cd res-downloader

# 2. 安装前端依赖
cd frontend
npm install
cd ..

# 3. 开发模式运行
wails dev
```

### 编译打包

```bash
# 构建 Web 资源
cd frontend
npm run build
cd ..

# 打包桌面应用(当前平台)
wails build

# 打包指定平台
wails build -platform windows/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64
```

---

## 扩展开发指南

### 1. 新增平台插件

**场景**: 需要支持新的视频/社交平台

**步骤**:

1. **创建插件文件**:

```go
// core/plugins/plugin.example.com.go
package plugins

import (
    "github.com/elazarl/goproxy"
    "net/http"
    "res-downloader/core/shared"
)

type ExamplePlugin struct {
    bridge *shared.Bridge
}

func (p *ExamplePlugin) SetBridge(bridge *shared.Bridge) {
    p.bridge = bridge
}

func (p *ExamplePlugin) Domains() []string {
    return []string{"example.com"}
}

func (p *ExamplePlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    // 可选: 修改请求
    return nil, nil
}

func (p *ExamplePlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    if resp.StatusCode != 200 {
        return nil
    }

    // 1. 检查 Content-Type
    contentType := resp.Header.Get("Content-Type")
    classify, suffix := p.bridge.TypeSuffix(contentType)

    if classify != "video" && classify != "audio" {
        return nil
    }

    // 2. 检查是否已标记
    url := resp.Request.URL.String()
    urlSign := shared.Md5(url)
    if p.bridge.MediaIsMarked(urlSign) {
        return nil
    }

    // 3. 构建资源信息
    mediaInfo := shared.MediaInfo{
        Id:          shared.GenerateID(),
        Url:         url,
        UrlSign:     urlSign,
        Classify:    classify,
        Suffix:      suffix,
        ContentType: contentType,
        Status:      shared.DownloadStatusReady,
    }

    // 4. 提取额外信息(如封面、描述)
    // ... 根据平台特定逻辑处理

    // 5. 标记并发送
    p.bridge.MarkMedia(urlSign)
    p.bridge.Send("newResources", mediaInfo)

    return nil
}
```

2. **注册插件**:

```go
// core/proxy.go 的 init() 函数
func init() {
    ps := []shared.Plugin{
        &plugins.QqPlugin{},
        &plugins.DefaultPlugin{},
        &plugins.ExamplePlugin{},  // 新增
    }
    // ... 其余代码
}
```

3. **测试验证**:
   - 启动代理
   - 访问目标网站
   - 检查资源是否出现在列表中

### 2. 自定义资源过滤规则

**场景**: 只拦截特定域名或路径的资源

**方案 1: 修改 RuleSet** (core/rule.go)

```go
func (r *RuleSet) shouldMitm(host string) bool {
    // 添加自定义逻辑
    if strings.Contains(host, "ads.example.com") {
        return false  // 不拦截广告域名
    }

    if strings.HasSuffix(host, ".example.com") {
        return true  // 拦截所有 example.com 子域名
    }

    // ... 默认逻辑
}
```

**方案 2: 在插件中过滤**

```go
func (p *ExamplePlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 只拦截特定路径
    if !strings.Contains(resp.Request.URL.Path, "/video/") {
        return nil
    }

    // 只拦截特定大小
    if resp.ContentLength < 1024*1024 {  // 小于 1MB
        return nil
    }

    // ... 处理逻辑
}
```

### 3. 修改下载行为

**场景**: 添加下载限速、自定义重试策略

**修改 downloader.go**:

```go
// 添加限速器
import (
    "golang.org/x/time/rate"
)

type FileDownloader struct {
    // ... 原有字段
    limiter *rate.Limiter  // 限速器
}

func (fd *FileDownloader) init() error {
    // ... 原有逻辑

    // 创建限速器(例如: 5MB/s)
    fd.limiter = rate.NewLimiter(rate.Limit(5*1024*1024), 1024*1024)

    return nil
}

func (fd *FileDownloader) doDownloadTask(progressChan chan ProgressChan, task *DownloadTask) error {
    // ... 原有逻辑

    buf := make([]byte, 32*1024)
    for {
        n, err := resp.Body.Read(buf)
        if n > 0 {
            // 限速
            if err := fd.limiter.WaitN(context.Background(), n); err != nil {
                return err
            }

            // ... 写入文件
        }
        // ...
    }
}
```

### 4. 扩展前端功能

**场景**: 添加批量重命名、去重功能

**步骤**:

1. **在 index.vue 中添加方法**:

```typescript
const batchRename = (prefix: string) => {
  data.value.forEach((item, index) => {
    if (checkedRowKeysValue.value.includes(item.Id)) {
      item.Description = `${prefix}_${index + 1}`
    }
  })
  cacheData()
}

const deduplicate = () => {
  const seen = new Set()
  data.value = data.value.filter(item => {
    if (seen.has(item.UrlSign)) {
      return false
    }
    seen.add(item.UrlSign)
    return true
  })
  cacheData()
}
```

2. **添加 UI 按钮**:

```vue
<NButton tertiary type="warning" @click.stop="showRenameDialog">
  批量重命名
</NButton>
<NButton tertiary type="success" @click.stop="deduplicate">
  去重
</NButton>
```

### 5. 修改默认配置

**场景**: 修改默认下载目录、线程数等

**修改 config.go**:

```go
func initConfig() *Config {
    defaultConfig := &Config{
        // ... 其他配置

        SaveDirectory: getDefaultDownloadDir(),
        TaskNumber:    runtime.NumCPU() * 2,  // 默认线程数
        DownNumber:    3,                      // 默认并发下载数

        // 自定义默认值
        FilenameTime:  false,  // 不添加时间戳
        UseHeaders:    "all",  // 使用所有 Header
    }

    // ...
}
```

### 6. 添加新的资源类型

**场景**: 支持新的 MIME 类型

**修改 config.go**:

```go
func getDefaultMimeMap() map[string]MimeInfo {
    return map[string]MimeInfo{
        // ... 原有类型

        // 新增类型
        "application/epub+zip":  {Type: "ebook", Suffix: ".epub"},
        "application/zip":       {Type: "archive", Suffix: ".zip"},
        "application/x-rar":     {Type: "archive", Suffix: ".rar"},
    }
}
```

---

## 常见问题

### 1. HTTPS 资源无法拦截?

**原因**: 未正确安装 CA 证书或域名未加入 MITM 规则

**解决**:
- 确保已安装证书:`/api/install`
- 检查 `rule.go` 中的 `shouldMitm` 逻辑
- 手动将域名加入配置的 `Rule` 字段

### 2. 下载速度慢?

**原因**: 默认线程数可能不足

**解决**:
- 调整 `config.TaskNumber` (建议 CPU 核心数 × 4)
- 启用 `DownloadProxy` 使用上游代理
- 检查网络带宽

### 3. 前端收不到后端事件?

**原因**: Wails 事件未正确绑定

**解决**:
- 确保调用了 `eventStore.init()`
- 检查事件类型是否匹配
- 查看浏览器控制台错误日志

### 4. 如何调试插件?

**方法**:
1. 添加日志输出:
   ```go
   globalLogger.Info().Msgf("Plugin processing: %s", host)
   ```
2. 使用 Wireshark 抓包对比
3. 启用 goproxy 的详细模式:
   ```go
   p.Proxy.Verbose = true
   ```

### 5. 编译失败?

**常见问题**:
- Go 版本过低 → 升级到 1.18+
- 依赖下载失败 → 配置 Go Proxy: `go env -w GOPROXY=https://goproxy.cn`
- Wails 版本不匹配 → 更新 Wails CLI

---

## 进阶主题

### 性能优化

1. **减少内存占用**: 控制资源列表大小,定期清理已完成任务
2. **并发控制**: 使用信号量限制协程数量
3. **缓存优化**: 对频繁访问的配置使用内存缓存

### 安全加固

1. **证书验证**: 检查客户端证书有效性
2. **数据加密**: 敏感配置使用 AES 加密存储
3. **权限控制**: 文件操作前检查权限

### 跨平台适配

**Windows**: 使用注册表设置代理
**macOS**: 使用 `networksetup` 命令
**Linux**: 修改环境变量或配置文件

核心代码位于 `core/system_*.go`

---

## 相关资源

- [Wails 官方文档](https://wails.io/docs/introduction)
- [goproxy 项目地址](https://github.com/elazarl/goproxy)
- [Vue 3 官方文档](https://cn.vuejs.org/)
- [Naive UI 组件库](https://www.naiveui.com/)

---

**最后更新**: 2026-01-28
**文档版本**: v3.1.3
