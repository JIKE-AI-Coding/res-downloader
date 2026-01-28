# 常见问题排查指南

> 本文档收集了 res-downloader 二次开发过程中的常见问题及解决方案。

---

## 目录

- [编译问题](#编译问题)
- [运行时问题](#运行时问题)
- [功能问题](#功能问题)
- [性能问题](#性能问题)
- [平台特定问题](#平台特定问题)

---

## 编译问题

### 问题 1: Wails 构建失败

**错误信息**:

```
error: failed to solve: executor failed running [/bin/sh -c wails build]
```

**原因**:
- Go 版本过低
- 依赖下载失败
- 网络问题

**解决方案**:

```bash
# 1. 检查 Go 版本
go version  # 需要 >= 1.18

# 2. 清理缓存
wails build -clean
go clean -cache -modcache

# 3. 配置 Go Proxy
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 4. 重新构建
wails dev
```

---

### 问题 2: 前端依赖安装失败

**错误信息**:

```
npm ERR! code ERESOLVE
npm ERR! ERESOLVE unable to resolve dependency tree
```

**原因**:
- Node 版本不兼容
- npm 版本过低

**解决方案**:

```bash
# 1. 检查 Node 版本
node -v  # 需要 >= 16

# 2. 清理 node_modules
cd frontend
rm -rf node_modules package-lock.json

# 3. 使用 Legacy Peer Dependencies
npm install --legacy-peer-deps

# 或使用 yarn
yarn install
```

---

### 问题 3: CGO 交叉编译失败

**错误信息**:

```
# github.com/webview/webview
./webview.go:10:10: fatal error: webview.h: No such file or directory
```

**原因**:
- 缺少交叉编译工具链
- CGO 未启用

**解决方案**:

```bash
# macOS → Windows
brew install mingw-w64
export CC=x86_64-w64-mingw32-gcc
export CGO_ENABLED=1

# Linux
sudo apt-get install build-essential

# 验证
go env CGO_ENABLED  # 应该输出 1
```

---

## 运行时问题

### 问题 4: HTTPS 资源无法拦截

**症状**:
- HTTP 资源正常
- HTTPS 资源不出现

**原因**:
- CA 证书未安装
- 域名未加入 MITM 规则

**解决方案**:

```bash
# 1. 检查证书安装
# - Windows: 控制面板 → 管理 → 证书 → 受信任的根证书颁发机构
# - macOS: 钥匙串访问 → 系统 → 证书
# - Linux: /usr/local/share/ca-certificates/

# 2. 重新安装证书
# 打开应用,会自动触发安装流程
# 或手动下载: http://127.0.0.1:8899/api/cert

# 3. 检查规则配置
# core/rule.go
func (r *RuleSet) shouldMitm(host string) bool {
    // 添加日志
    fmt.Printf("Checking host: %s, result: %v\n", host, result)
    // ...
}

# 4. 临时允许所有域名
# 在设置中将 Rule 字段改为 "*"
```

---

### 问题 5: 系统代理设置失败

**错误信息**:

```
Error: permission denied
```

**原因**:
- 缺少管理员权限
- macOS/Linux 需要密码

**解决方案**:

```bash
# Windows
# 右键应用 → 以管理员身份运行

# macOS/Linux
# 应用会弹出密码输入框
# 或在终端中运行:
sudo res-downloader
```

---

### 问题 6: 前端收不到后端事件

**症状**:
- 资源被后端捕获
- 前端列表不更新

**原因**:
- Wails 事件未绑定
- 事件类型不匹配

**解决方案**:

```typescript
// 1. 检查事件监听器
eventStore.init()  // 确保已初始化

// 2. 添加调试日志
eventStore.addHandle({
    type: "newResources",
    event: (res: appType.MediaInfo) => {
        console.log("Received resource:", res)
        data.value.push(res)
    }
})

// 3. 检查后端发送
// core/http.go
func (h *HttpServer) send(t string, data interface{}) {
    fmt.Printf("Sending event: %s, data: %v\n", t, data)
    runtime.EventsEmit(appOnce.ctx, "event", string(jsonData))
}

// 4. 验证 JSON 序列化
jsonData, err := json.Marshal(map[string]interface{}{
    "type": t,
    "data": data,
})
if err != nil {
    fmt.Printf("JSON marshal error: %v\n", err)
}
```

---

## 功能问题

### 问题 7: 特定网站资源无法捕获

**原因**:
- 网站使用 WebSocket
- 网站使用分段下载
- 网站检测代理

**解决方案**:

```go
// 方案 1: 添加自定义插件
// core/plugins/plugin.example.com.go

func (p *ExamplePlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    // 修改请求头,绕过检测
    r.Header.Set("Proxy-Connection", "")
    r.Header.Del("Via")
    return nil, nil
}

func (p *ExamplePlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 检查 WebSocket
    if resp.Header.Get("Upgrade") == "websocket" {
        // WebSocket 暂不支持
        return nil
    }

    // 检查分段下载
    if resp.Header.Get("Accept-Ranges") == "bytes" {
        // 仍然捕获,但只能下载第一段
        // 完整实现需要合并分段
    }

    return nil
}
```

---

### 问题 8: 下载速度慢

**原因**:
- 线程数设置过低
- 服务器限速
- 网络带宽不足

**解决方案**:

```go
// 1. 调整线程数
// core/config.go
defaultConfig := &Config{
    TaskNumber: runtime.NumCPU() * 4,  // 增加到 4 倍
}

// 2. 启用上游代理
// 在设置中配置 UpstreamProxy
// 例如: socks5://127.0.0.1:1080

// 3. 添加限速逻辑
// core/downloader.go
type FileDownloader struct {
    // ...
    rateLimiter *rate.Limiter
}

func (fd *FileDownloader) doDownloadTask(...) error {
    // 添加限速
    if fd.rateLimiter != nil {
        _ = fd.rateLimiter.WaitN(ctx, n)
    }
    // ...
}
```

---

### 问题 9: 微信视频号视频无法下载

**原因**:
- 视频加密
- 需要解密密钥

**解决方案**:

```javascript
// 前端解密示例
// frontend/src/assets/js/decrypt.js

export function getDecryptionArray(key) {
    // key 是 16 字节的密钥
    // 返回用于 XOR 解密的字节数组
    const keyBytes = atob(key)
    const result = new Uint8Array(16)
    for (let i = 0; i < 16; i++) {
        result[i] = keyBytes.charCodeAt(i)
    }
    return result
}

// 使用
const decodeStr = uint8ArrayToBase64(getDecryptionArray(row.DecodeKey))
appApi.download({...row, decodeStr})
```

---

## 性能问题

### 问题 10: 内存占用过高

**原因**:
- 资源列表未清理
- 并发下载过多

**解决方案**:

```go
// 1. 定期清理已完成任务
// core/resource.go
func (r *Resource) cleanup() {
    r.tasks.Range(func(key, value interface{}) bool {
        d := value.(*FileDownloader)
        if d.TotalSize > 0 && d.File != nil {
            d.File.Close()
            r.tasks.Delete(key)
        }
        return true
    })
}

// 2. 限制资源列表大小
// core/resource.go
const maxResources = 1000

func (r *Resource) markMedia(key string) {
    // 检查数量
    count := 0
    r.mediaMark.Range(func(_, _) bool {
        count++
        return count < maxResources
    })

    if count >= maxResources {
        r.clear()  // 清空列表
    }

    r.mediaMark.Store(key, true)
}
```

---

### 问题 11: CPU 占用过高

**原因**:
- 无限循环
- 频繁的文件 I/O

**解决方案**:

```go
// 1. 添加限流
import "time"

func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 避免处理过快
    time.Sleep(10 * time.Millisecond)
    // ...
}

// 2. 使用缓存
type CachedResponse struct {
    Body       []byte
    Header     http.Header
    StatusCode int
}

var responseCache = sync.Map{}

func (p *YourPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    url := resp.Request.URL.String()

    // 检查缓存
    if cached, ok := responseCache.Load(url); ok {
        // 返回缓存的响应
        return cached.(*CachedResponse).toResponse()
    }

    // 处理并缓存
    // ...
}
```

---

## 平台特定问题

### Windows

#### 问题 12: 杀毒软件误报

**原因**:
- 未签名的可执行文件
- 包含代理功能

**解决方案**:
- 申请代码签名证书
- 添加白名单
- 用户手动排除

---

#### 问题 13: Win7 兼容性

**解决方案**:
- 使用 Electron 旧版 (分支 `old`)
- 或升级到 Windows 10+

---

### macOS

#### 问题 14: 公证失败

**错误信息**:

```
Error: Unable to notarize
```

**解决方案**:

```bash
# 1. 检查 Apple ID
# 需要启用 App-specific password
# https://appleid.apple.com/account/manage

# 2. 检查 Team ID
# https://developer.apple.com/account

# 3. 手动重试
xcrun notarytool submit res-downloader.app.zip \
  --apple-id "your@email.com" \
  --password "app-specific-password" \
  --team-id "TEAM_ID" \
  --wait

# 4. 查看详细日志
xcrun notarytool log \
  --apple-id "your@email.com" \
  --password "app-specific-password" \
  --team-id "TEAM_ID" \
  SUBMISSION_ID
```

---

#### 问题 15: ARM64 架构支持

**解决方案**:

```bash
# 1. 使用 Rosetta 运行 x86_64 版本
softwareupdate --install-rosetta

# 2. 编译原生 ARM64 版本
GOARCH=arm64 wails build

# 3. 创建 Universal Binary
lipo -create \
  -output res-downloader-universal \
  res-downloader-x86_64 \
  res-downloader-arm64
```

---

### Linux

#### 问题 16: 缺少运行时依赖

**错误信息**:

```
error while loading shared libraries: libwebkit2gtk-4.0.so.37
```

**解决方案**:

```bash
# Ubuntu/Debian
sudo apt-get install libwebkit2gtk-4.0-37 libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install webkit2gtk3

# Arch Linux
sudo pacman -S webkit2gtk

# 静态编译(推荐)
CGO_ENABLED=1 go build \
  -ldflags="-linkmode external -extldflags '-static-libgcc'"
```

---

## 调试技巧

### 1. 启用详细日志

```go
// core/logger.go
type Logger struct {
    level string  // "debug", "info", "warn", "error"
}

// 在代码中添加日志
globalLogger.Debug().Msgf("Processing URL: %s", url)
globalLogger.Info().Msgf("Download started: %s", filename)
globalLogger.Warn().Msg("Retry limit reached")
globalLogger.Err().Msgf("Download failed: %v", err)
```

---

### 2. 网络抓包

```bash
# 使用 tcpdump
sudo tcpdump -i any -w output.pcap host example.com

# 使用 Wireshark
# 打开 output.pcap 进行分析
```

---

### 3. 性能分析

```bash
# CPU 性能分析
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof
go tool pprof mem.prof

# 生成火焰图
go tool pprof -http=:8080 cpu.prof
```

---

### 4. pprof 集成

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // 应用主逻辑
    // ...
}
```

访问 `http://localhost:6060/debug/pprof/` 查看性能数据。

---

**最后更新**: 2026-01-28
