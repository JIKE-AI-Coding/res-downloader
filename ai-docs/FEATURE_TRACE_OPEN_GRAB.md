# 开启抓取功能 - 端到端分析

> 本文档使用 `trace-feature-end-to-end` 技能生成
> 分析目标：点击"开启抓取"后的完整处理逻辑

---

## 【功能点概述】

### 用户视角功能描述
用户点击"开启抓取"按钮后，应用会：
- 开启系统代理（将系统网络流量导向本地代理服务器）
- 启动 HTTP/HTTPS 流量拦截与嗅探
- 自动捕获符合条件的网络资源（视频、音频、图片等）
- 将捕获的资源实时显示在前端列表中

### 触发条件
- **触发方式**：用户点击主页顶部的"开启抓取"按钮
- **权限要求**：macOS/Linux 需要 sudo 密码，Windows 可能需要管理员权限
- **前置条件**：必须已安装根证书（首次使用会自动触发安装流程）

### 输入与输出
- **输入**：用户点击操作 + 系统密码（macOS/Linux）
- **输出**：
  - 系统代理状态变更
  - 前端按钮变为"关闭抓取"（带红色动画指示器）
  - 实时捕获的资源显示在列表中

---

## 【前端实现链路】

### 页面 / 组件

#### 入口组件
**文件**: `frontend/src/views/index.vue:5-11`

```vue
<NButton v-if="isProxy" secondary type="primary" @click.stop="close">
  <span class="inline-block w-1.5 h-1.5 bg-red-600 rounded-full mr-1 animate-pulse"></span>
  {{ t("index.close_grab") }}{{ data.length > 0 ? `&nbsp;${t('index.total_resources', {count: data.length})}` : '' }}
</NButton>
<NButton v-else tertiary type="tertiary" @click.stop="open">
  {{ t("index.open_grab") }}{{ data.length > 0 ? `&nbsp;${t('index.total_resources', {count: data.length})}` : '' }}
</NButton>
```

**关键点**：
- 通过 `v-if="isProxy"` 切换按钮显示
- "开启抓取" 绑定 `@click.stop="open"` 事件
- "关闭抓取" 绑定 `@click.stop="close"` 事件
- 代理开启时显示红色呼吸灯动画（`animate-pulse`）

#### 事件处理函数
**文件**: `frontend/src/views/index.vue:872-885`

```typescript
const open = () => {
  isOpenProxy = true
  store.openProxy().then((res: appType.Res) => {
    if (res.code === 1) {
      return
    }

    // macOS/Linux 需要密码授权
    if (["darwin", "linux"].includes(store.envInfo.platform)) {
      showPassword.value = true
    } else {
      window.$message?.error(res.message)
    }
  })
}
```

**关闭函数**: `frontend/src/views/index.vue:887-889`
```typescript
const close = () => {
  store.unsetProxy()
}
```

### 关键事件与状态

#### 状态管理流程
```
用户点击 → open() → store.openProxy()
  ↓
Pinia Store (stores/index.ts:72-74)
  ↓
appApi.openSystemProxy()
  ↓
HTTP POST /api/proxy-open
  ↓
后端处理 → 返回响应
  ↓
handleProxy() (stores/index.ts:80-86)
  ↓
更新 isProxy 状态 → UI 自动刷新
```

#### 状态定义
**文件**: `frontend/src/stores/index.ts:46`

```typescript
const isProxy = ref(false)
```

**计算属性**: `frontend/src/views/index.vue:173-175`
```typescript
const isProxy = computed(() => {
  return store.isProxy
})
```

### 关键代码位置

| 层级 | 文件路径 | 行号 | 功能 |
|------|---------|------|------|
| **UI 层** | `frontend/src/views/index.vue` | 5-11 | 按钮渲染与绑定 |
| **UI 层** | `frontend/src/views/index.vue` | 872-885 | 开启代理事件处理 |
| **UI 层** | `frontend/src/views/index.vue` | 887-889 | 关闭代理事件处理 |
| **状态层** | `frontend/src/stores/index.ts` | 72-86 | openProxy/unsetProxy 实现 |
| **API 层** | `frontend/src/api/app.ts` | 17-22 | openSystemProxy API 定义 |
| **API 层** | `frontend/src/api/app.ts` | 23-28 | unsetSystemProxy API 定义 |

---

## 【接口调用】

### API 路径

#### 开启代理
```
POST /api/proxy-open
```

**文件**: `frontend/src/api/app.ts:17-22`
```typescript
openSystemProxy() {
  return request({
    url: 'api/proxy-open',
    method: 'post',
  })
}
```

#### 关闭代理
```
POST /api/proxy-unset
```

**文件**: `frontend/src/api/app.ts:23-28`
```typescript
unsetSystemProxy() {
  return request({
    url: 'api/proxy-unset',
    method: 'post',
  })
}
```

### 请求 / 响应关键字段

#### 请求
- **请求体**: 无需参数
- **Headers**: 由 `request` 拦截器自动添加

#### 响应
**成功响应**:
```json
{
  "code": 1,
  "message": "ok",
  "data": {
    "value": true
  }
}
```

**失败响应**:
```json
{
  "code": 0,
  "message": "错误详情（如权限不足、命令执行失败等）",
  "data": {
    "value": false
  }
}
```

#### 状态码说明
- `code: 1`: 成功
- `code: 0`: 失败
- `data.value`: 当前代理状态（true/false）

---

## 【后端处理】

### 入口处理（HTTP 层）

**文件**: `core/http.go:254-265`

```go
func (h *HttpServer) openSystemProxy(w http.ResponseWriter, r *http.Request) {
    err := appOnce.OpenSystemProxy()
    if err != nil {
        h.error(w, err.Error(), respData{
            "value": appOnce.IsProxy,
        })
        return
    }
    h.success(w, respData{
        "value": appOnce.IsProxy,
    })
}
```

**文件**: `core/http.go:267-278`
```go
func (h *HttpServer) unsetSystemProxy(w http.ResponseWriter, r *http.Request) {
    err := appOnce.UnsetSystemProxy()
    if err != nil {
        h.error(w, err.Error(), respData{
            "value": appOnce.IsProxy,
        })
        return
    }
    h.success(w, respData{
        "value": appOnce.IsProxy,
    })
}
```

### 核心业务逻辑（应用层）

**文件**: `core/app.go:156-166`

```go
func (a *App) OpenSystemProxy() error {
    if a.IsProxy {
        return nil  // 避免重复开启
    }
    err := systemOnce.setProxy()
    if err == nil {
        a.IsProxy = true
        return nil
    }
    return err
}
```

**文件**: `core/app.go:168-178`

```go
func (a *App) UnsetSystemProxy() error {
    if !a.IsProxy {
        return nil  // 避免重复关闭
    }
    err := systemOnce.unsetProxy()
    if err == nil {
        a.IsProxy = false
        return nil
    }
    return err
}
```

### 系统代理设置（平台相关）

#### macOS 实现
**文件**: `core/system_darwin.go:61-88`

```go
func (s *SystemSetup) setProxy() error {
    services, err := s.getNetworkServices()
    if err != nil {
        return err
    }

    isSuccess := false
    var errs strings.Builder
    for _, serviceName := range services {
        commands := [][]string{
            {"networksetup", "-setwebproxy", serviceName, "127.0.0.1", globalConfig.Port},
            {"networksetup", "-setsecurewebproxy", serviceName, "127.0.0.1", globalConfig.Port},
        }
        for _, cmd := range commands {
            if output, err := s.runCommand(cmd); err != nil {
                errs.WriteString(fmt.Sprintf("cmd: %v\noutput: %s\nerr: %s\n", cmd, output, err))
            } else {
                isSuccess = true
            }
        }
    }

    if isSuccess {
        return nil
    }

    return fmt.Errorf("failed to set proxy for any active network service, errs:%s", errs)
}
```

**关键步骤**：
1. 获取所有活动的网络服务（Wi-Fi、Ethernet 等）
2. 为每个服务设置 HTTP 代理（`-setwebproxy`）
3. 为每个服务设置 HTTPS 代理（`-setsecurewebproxy`）
4. 所有命令通过 `sudo` 执行（需要密码）

#### Windows 实现
**文件**: `core/system_windows.go`

使用 Windows Registry 修改代理设置

#### Linux 实现
**文件**: `core/system_linux.go`

使用 `gsettings` 或网络管理工具配置代理

### 代理服务器启动

**文件**: `core/app.go:129-132`

```go
func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    go httpServerOnce.run()  // 在 goroutine 中启动 HTTP 服务器
}
```

**文件**: `core/http.go:37-54`

```go
func (h *HttpServer) run() {
    listener, err := net.Listen("tcp", globalConfig.Host+":"+globalConfig.Port)
    if err != nil {
        globalLogger.Err(err)
        log.Fatalf("Service cannot start: %v", err)
    }
    fmt.Println("Service started, listening http://" + globalConfig.Host + ":" + globalConfig.Port)
    if err1 := http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Host == "127.0.0.1:"+globalConfig.Port && HandleApi(w, r) {
            // 处理 API 请求
        } else {
            proxyOnce.Proxy.ServeHTTP(w, r)  // 转发到代理
        }
    })); err1 != nil {
        globalLogger.Err(err1)
        fmt.Printf("Service startup exception: %v", err1)
    }
}
```

**关键点**：
- 监听 `0.0.0.0:8899`（默认）
- 区分 API 请求和代理流量
- API 请求：`r.Host == "127.0.0.1:8899"`
- 代理流量：转发到 `goproxy` 处理

### 资源捕获逻辑

#### MITM (中间人攻击) 拦截
**文件**: `core/proxy.go:84-92`

```go
p.Proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
    if ruleOnce.shouldMitm(host) {  // 检查规则是否允许拦截
        return goproxy.MitmConnect, host
    }
    return goproxy.OkConnect, host
})
```

**流程**：
1. 收到 HTTPS CONNECT 请求
2. 检查域名是否匹配规则
3. 匹配：使用 MITM 拦截（解密流量）
4. 不匹配：直接转发（不解密）

#### 插件处理流程
**文件**: `core/proxy.go:144-157`

```go
func (p *Proxy) httpRequestEvent(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    plugin := p.matchPlugin(r.Host)  // 匹配域名对应的插件
    if plugin != nil {
        newReq, newResp := plugin.OnRequest(r, ctx)
        if newResp != nil {
            return newReq, newResp
        }
        if newReq != nil {
            return newReq, nil
        }
    }
    return pluginRegistry["default"].OnRequest(r, ctx)
}
```

**文件**: `core/proxy.go:159-173`

```go
func (p *Proxy) httpResponseEvent(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    if resp == nil || resp.Request == nil {
        return resp
    }

    plugin := p.matchPlugin(resp.Request.Host)
    if plugin != nil {
        newResp := plugin.OnResponse(resp, ctx)
        if newResp != nil {
            return newResp
        }
    }

    return pluginRegistry["default"].OnResponse(resp, ctx)
}
```

#### 默认插件资源嗅探
**文件**: `core/plugins/plugin.default.go:50-87`

```go
// 1. 检查响应类型
contentType := resp.Header.Get("Content-Type")
// ...

// 2. 过滤资源类型
ok, _ = p.bridge.GetResType(mimeType)
if !ok {
    return resp
}

// 3. 过滤文件大小
contentLength := resp.ContentLength
if contentLength > 0 && contentLength < globalConfig.Quality {
    return resp
}

// 4. 防重复
urlSign := shared.GetMd5Str(req.URL.String())
if p.bridge.MediaIsMarked(urlSign) {
    return resp
}

// 5. 构造资源信息
res := shared.MediaInfo{
    Id:         shared.GetUuid(),
    Url:        req.URL.String(),
    UrlSign:    urlSign,
    Domain:     req.URL.Host,
    Classify:   classify,
    Size:       contentLength,
    Status:     "ready",
    Headers:    headers,
}

// 6. 标记并发送到前端
p.bridge.MarkMedia(urlSign)
go func(res shared.MediaInfo) {
    p.bridge.Send("newResources", res)
}(res)

return resp
```

---

## 【数据层】

### 数据模型

#### MediaInfo 结构体
**定义位置**: `core/shared/media.go`

```go
type MediaInfo struct {
    Id          string            `json:"Id"`
    Url         string            `json:"Url"`
    UrlSign     string            `json:"UrlSign"`
    CoverUrl    string            `json:"CoverUrl"`
    Domain      string            `json:"Domain"`
    Classify    string            `json:"Classify"`
    Size        int64             `json:"Size"`
    Description string            `json:"Description"`
    Status      string            `json:"Status"`
    SavePath    string            `json:"SavePath"`
    Headers     map[string]string `json:"Headers"`
    Extension   string            `json:"Extension"`
    Filename    string            `json:"Filename"`
    DecodeKey   []byte            `json:"DecodeKey,omitempty"`
}
```

#### 字段说明
| 字段 | 类型 | 说明 |
|------|------|------|
| `Id` | string | 唯一标识符（UUID） |
| `Url` | string | 资源 URL |
| `UrlSign` | string | URL 签名（MD5，用于去重） |
| `Domain` | string | 域名 |
| `Classify` | string | 资源类型（video/audio/image/m3u8/live 等） |
| `Size` | int64 | 文件大小（字节） |
| `Status` | string | 状态（ready/running/done/error） |
| `SavePath` | string | 保存路径 |
| `DecodeKey` | []byte | 解密密钥（如微信视频） |

### 关键字段

#### 代理状态
**文件**: `core/app.go:28`
```go
type App struct {
    // ...
    IsProxy     bool   `json:"IsProxy"`
    // ...
}
```

#### 配置端口
**文件**: `core/config.go`
```go
type Config struct {
    Host string `json:"Host"`
    Port string `json:"Port"`
    // 默认值：Host="0.0.0.0", Port="8899"
}
```

### 数据流转

```
┌─────────────────────────────────────────────────────────────────┐
│                        完整数据流                                 │
└─────────────────────────────────────────────────────────────────┘

1. 网络流量层
   浏览器请求 → 系统代理 → 127.0.0.1:8899

2. 代理服务器层 (proxy.go)
   goproxy 接收请求 → HandleConnectFunc 判断
                           ↓
                   shouldMitm(host) 规则检查
                           ↓
              ┌────────────┴────────────┐
              ↓                         ↓
        MITM 解密                   直接转发
              ↓                         ↓
    httpRequestEvent            原始流量
              ↓
    httpResponseEvent
              ↓
    插件处理 (plugin.default.go)
              ↓
3. 资源构造层
   检查 MIME Type
   检查文件大小
   检查是否重复 (MediaIsMarked)
   构造 MediaInfo
              ↓
4. 事件发送层
   MarkMedia(urlSign) - 标记
   Send("newResources", MediaInfo) - 发送
              ↓
5. Wails 运行时层
   runtime.EventsEmit(ctx, "event", JSON)
              ↓
6. 前端事件层
   EventsOn("event") → EventStore
              ↓
   handles["newResources"](data)
              ↓
7. 前端数据层
   data.value.push(res) 或 unshift(res)
              ↓
8. 本地缓存
   localStorage.setItem("resources-data", JSON.stringify(data))
              ↓
9. UI 展示
   NDataTable 渲染列表
```

---

## 【修改建议】

### 推荐修改点

#### 1. 修改资源过滤规则

**位置**: `core/rule.go` (需检查具体实现)

**适用场景**:
- 添加/删除需要拦截的域名
- 修改匹配规则（正则表达式、通配符等）

**影响范围**: 影响所有流量的 MITM 决策

**示例**:
```go
// 假设的规则添加
func (r *RuleSet) shouldMitm(host string) bool {
    // 白名单模式
    if r.isWhitelistMode {
        return r.whitelist[host]
    }
    // 黑名单模式
    return !r.blacklist[host]
}
```

**风险提示**:
- 过于宽泛的规则会导致性能问题
- 拦截敏感站点可能导致安全风险

#### 2. 新增资源类型支持

**位置**: `core/plugins/` 目录下的插件文件

**步骤**:

1. **修改配置文件** (`core/config.go` 或配置界面):
```json
{
  "MimeMap": {
    "application/x-new-type": {
      "Type": "newtype",
      "Extension": ".newext"
    }
  }
}
```

2. **修改插件逻辑** (`core/plugins/plugin.default.go`):
```go
// 在 OnResponse 中添加新的 MIME 类型判断
if strings.Contains(mimeType, "application/x-new-type") {
    classify = "newtype"
    // ... 处理逻辑
}
```

3. **更新前端映射** (`frontend/src/views/index.vue:203-214`):
```typescript
const classifyAlias: { [key: string]: any } = {
    // ...
    newtype: computed(() => t("index.newtype"))
}
```

**影响范围**:
- 资源识别逻辑
- 前端显示
- 下载处理（如果需要特殊处理）

#### 3. 修改抓取触发条件

**位置**: `core/plugins/plugin.default.go:50-87`

**关键参数**:

| 参数 | 位置 | 说明 |
|------|------|------|
| MIME 类型 | 第61-66行 | `p.bridge.GetResType(mimeType)` |
| 文件大小 | 第69行 | `contentLength < globalConfig.Quality` |
| URL 模式 | 需查看 | 可能需要添加 URL 匹配逻辑 |

**示例 - 添加 URL 模式过滤**:
```go
// 在第69行后添加
if !strings.Contains(req.URL.String(), "target-pattern") {
    return resp  // 不匹配的 URL 直接放行
}
```

**示例 - 添加域名白名单**:
```go
allowedDomains := []string{"example.com", "cdn.example.com"}
isAllowed := false
for _, domain := range allowedDomains {
    if strings.HasSuffix(req.URL.Host, domain) {
        isAllowed = true
        break
    }
}
if !isAllowed {
    return resp
}
```

#### 4. 修改资源去重策略

**位置**: `core/plugins/plugin.default.go:73-75`

**当前逻辑**:
```go
urlSign := shared.GetMd5Str(req.URL.String())
if p.bridge.MediaIsMarked(urlSign) {
    return resp  // 已存在，跳过
}
```

**可选优化**:
```go
// 1. 考虑查询参数的去重
urlWithoutQuery := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
urlSign := shared.GetMd5Str(urlWithoutQuery)

// 2. 添加时效性检查（允许重新抓取旧资源）
if p.bridge.MediaIsMarked(urlSign) {
    // 检查首次捕获时间
    firstSeen := p.bridge.GetMediaFirstSeen(urlSign)
    if time.Since(firstSeen) < 24*time.Hour {
        return resp
    }
    // 超过24小时，允许重新抓取
}
```

#### 5. 添加自定义资源处理

**位置**: 新建插件文件 `core/plugins/plugin.custom.go`

**模板**:
```go
package plugins

import (
    "net/http"
    "res-downloader/core/shared"
    "github.com/elazarl/goproxy"
)

type CustomPlugin struct {
    bridge *shared.Bridge
}

func (p *CustomPlugin) SetBridge(bridge *shared.Bridge) {
    p.bridge = bridge
}

func (p *CustomPlugin) Domains() []string {
    return []string{"custom-domain.com"}
}

func (p *CustomPlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    // 自定义请求处理逻辑
    return r, nil
}

func (p *CustomPlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // 自定义响应处理逻辑
    // 例如：提取特殊的资源链接、解密内容等
    return resp
}
```

**注册插件** (`core/proxy.go:26-31`):
```go
ps := []shared.Plugin{
    &plugins.QqPlugin{},
    &plugins.DefaultPlugin{},
    &plugins.CustomPlugin{},  // 添加这一行
}
```

### 影响范围分析

#### 高耦合点（修改需谨慎）

| 模块 | 文件 | 耦合原因 | 影响范围 |
|------|------|----------|----------|
| **代理服务器启动** | `core/app.go:131` | 在应用启动时自动启动 | 所有代理功能 |
| **规则引擎** | `core/rule.go` | 决定哪些流量被拦截 | 所有资源捕获 |
| **插件系统** | `core/proxy.go` | 插件覆盖特定域名 | 该域名下的所有资源 |
| **系统代理设置** | `core/system_*.go` | 涉及系统网络配置 | 整个操作系统网络 |
| **Wails 事件系统** | `core/http.go:111-121` | 前后端通信桥梁 | 所有实时更新 |

#### 中等耦合点

| 模块 | 文件 | 影响范围 |
|------|------|----------|
| **资源存储** | `core/resource.go` | 资源列表、去重 |
| **配置管理** | `core/config.go` | 全局行为 |
| **默认插件** | `core/plugins/plugin.default.go` | 未被特定插件覆盖的域名 |

#### 低耦合点（推荐修改）

| 模块 | 文件 | 建议修改类型 |
|------|------|--------------|
| **前端 UI** | `frontend/src/views/index.vue` | 样式、交互 |
| **前端状态** | `frontend/src/stores/index.ts` | 新增状态字段 |
| **自定义插件** | `core/plugins/plugin.*.go` | 新增域名特定逻辑 |
| **配置界面** | `frontend/src/components/` | 新增配置项 |

### 风险提示

#### ⚠️ 系统代理操作风险

**macOS/Linux**:
- 需要 `sudo` 密码（通过 `Password` 组件收集）
- 密码通过 `systemOnce.Password` 缓存
- 执行失败需手动恢复网络设置

**Windows**:
- 可能需要管理员权限
- 修改注册表，需谨慎操作
- 失败可能导致网络异常

**建议**:
- 在测试环境充分验证
- 提供回滚机制
- 记录详细日志

#### ⚠️ 证书安装风险

**问题**:
- 证书安装失败会导致 HTTPS 嗅探失效
- 错误的证书可能导致安全警告

**相关代码**: `core/app.go:143-154`

```go
func (a *App) installCert() (string, error) {
    out, err := systemOnce.installCert()
    if err != nil {
        globalLogger.Esg(err, out)
        return out, err
    } else {
        if err := a.lock(); err != nil {
            globalLogger.Err(err)
        }
    }
    return out, nil
}
```

**建议**:
- 首次使用强制安装证书
- 提供证书验证功能
- 支持证书更新

#### ⚠️ 代理状态同步风险

**问题**:
- 前后端状态不一致导致 UI 错误
- 重复开启/关闭导致资源泄漏

**当前保护**:
```go
// core/app.go:156-159
func (a *App) OpenSystemProxy() error {
    if a.IsProxy {
        return nil  // 避免重复开启
    }
    // ...
}
```

**建议**:
- 添加状态定时同步
- 提供"重置状态"功能
- 记录状态变更日志

#### ⚠️ 高并发场景风险

**当前处理**:
```go
// core/plugins/plugin.default.go:80-83
go func(res shared.MediaInfo) {
    p.bridge.Send("newResources", res)
}(res)
```

**潜在问题**:
- 大量资源同时到达可能导致事件积压
- goroutine 泄漏风险

**建议**:
- 使用带缓冲的 channel
- 限制并发 goroutine 数量
- 添加资源去重队列

#### ⚠️ 性能风险

**场景**:
- 规则过于复杂导致 MITM 判断延迟
- 插件处理耗时影响代理性能
- 大文件响应处理占用内存

**优化建议**:
1. **规则优化**:
   - 使用缓存（LRU Cache）
   - 预编译正则表达式
   - 分层规则（先简单后复杂）

2. **插件优化**:
   - 避免在 `OnRequest/OnResponse` 中执行耗时操作
   - 使用 goroutine 处理资源发送
   - 限制响应体读取大小

3. **内存优化**:
   - 流式处理大文件
   - 及时释放资源
   - 监控内存使用

---

## 【架构边界（不应直接修改）】

### ❌ 禁止修改的部分

| 模块 | 文件 | 原因 |
|------|------|------|
| **Wails 运行时** | `wailsjs/runtime` | 框架核心，修改会导致通信失败 |
| **事件发送机制** | `runtime.EventsEmit()` | 前后端通信基础设施 |
| **goproxy 核心逻辑** | `github.com/elazarl/goproxy` | 第三方库，不理解会导致代理失效 |
| **系统命令执行** | `runCommand()` 中的 sudo 逻辑 | 涉及安全风险 |

### ✅ 安全修改的部分

| 模块 | 文件 | 可修改内容 |
|------|------|-----------|
| **前端组件** | `frontend/src/views/index.vue` | UI 样式、交互逻辑 |
| **前端状态** | `frontend/src/stores/index.ts` | 新增状态、计算属性 |
| **API 封装** | `frontend/src/api/app.ts` | 新增 API 调用 |
| **插件实现** | `core/plugins/plugin.*.go` | 域名特定逻辑 |
| **配置文件** | `core/config.go` | 默认配置值 |
| **HTTP 处理** | `core/http.go` | 新增 API 端点 |

### ⚠️ 谨慎修改的部分

| 模块 | 文件 | 注意事项 |
|------|------|---------|
| **默认插件** | `core/plugins/plugin.default.go` | 影响所有未覆盖域名 |
| **规则引擎** | `core/rule.go` | 影响所有流量拦截 |
| **资源管理** | `core/resource.go` | 影响去重、存储逻辑 |
| **系统代理设置** | `core/system_*.go` | 影响系统网络配置 |

---

## 【调试与故障排查】

### 常见问题

#### 1. 点击"开启抓取"后没有反应

**排查步骤**:
1. 检查浏览器开发者工具 Console 是否有错误
2. 检查 Network 标签，看 `/api/proxy-open` 请求是否发送
3. 检查后端日志：`core/app.go` 和 `core/http.go`

**可能原因**:
- 后端服务未启动
- 端口被占用
- 权限不足

#### 2. 提示权限错误

**macOS/Linux**:
```bash
# 手动测试命令
sudo -S networksetup -setwebproxy Wi-Fi 127.0.0.1 8899
```

**Windows**:
- 以管理员身份运行应用

#### 3. 资源无法捕获

**检查清单**:
- [ ] 证书是否安装
- [ ] 系统代理是否生效（浏览器网络设置）
- [ ] 域名是否匹配规则
- [ ] MIME 类型是否在配置中
- [ ] 资源是否已被标记（去重）

**调试代码**:
```go
// 在 plugin.default.go 中添加日志
fmt.Printf("DEBUG: URL=%s, MIME=%s, Size=%d\n", req.URL.String(), mimeType, contentLength)
```

#### 4. 前端列表不更新

**检查**:
1. 事件是否发送：查看后端日志
2. 事件监听是否注册：`frontend/src/views/index.vue:530-540`
3. 数据是否添加：添加 `console.log` 检查

```typescript
// 在 index.vue 中添加调试
eventStore.addHandle({
  type: "newResources",
  event: (res: appType.MediaInfo) => {
    console.log("Received resource:", res)  // 调试日志
    if (store.globalConfig.InsertTail) {
      data.value.push(res)
    } else {
      data.value.unshift(res)
    }
    cacheData()
  }
})
```

---

## 【相关文档】

- [开发指南](./DEV_GUIDE.md) - 完整的开发文档
- [插件开发](./DEV_PLUGIN_DEV.md) - 自定义插件开发
- [API 参考](./DEV_API_REFERENCE.md) - 接口详细说明
- [故障排查](./DEV_TROUBLESHOOTING.md) - 常见问题解决

---

## 【总结】

"开启抓取"功能的完整链路涉及：

1. **前端交互**：按钮点击 → 状态管理 → API 调用
2. **后端处理**：HTTP 接口 → 应用逻辑 → 系统代理设置
3. **代理服务**：流量拦截 → MITM 解密 → 插件处理
4. **资源捕获**：规则过滤 → 插件嗅探 → 事件发送
5. **前端展示**：事件接收 → 数据更新 → UI 渲染

**关键修改点**：
- 新增资源类型：修改 `MimeMap` 配置
- 修改过滤规则：编辑插件逻辑
- 添加域名支持：新建插件文件

**高风险操作**：
- 系统代理设置（需权限）
- 证书管理（影响 HTTPS）
- 规则引擎（影响所有流量）

**建议**：
- 优先添加自定义插件而非修改默认插件
- 配置修改优于代码修改
- 充分测试后再部署到生产环境
