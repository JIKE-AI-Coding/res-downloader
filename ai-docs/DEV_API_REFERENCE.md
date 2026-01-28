# API 参考

> res-downloader 后端 API 接口完整文档,包含所有 HTTP 接口和 Wails 绑定方法。

---

## 目录

- [HTTP API](#http-api)
- [Wails 绑定方法](#wails-绑定方法)
- [前端事件](#前端事件)
- [数据结构](#数据结构)
- [错误码](#错误码)

---

## HTTP API

### 基础信息

- **Base URL**: `http://127.0.0.1:8899/api`
- **Content-Type**: `application/json`
- **响应格式**:

```json
{
  "code": 1,          // 1=成功, 0=失败
  "message": "ok",
  "data": {}
}
```

---

### 证书管理

#### 安装证书

**接口**: `POST /api/install`

**说明**: 首次运行时安装 CA 证书到系统信任库

**请求**: 无需参数

**响应**:

```json
{
  "code": 1,
  "message": "ok",
  "data": {
    "isPass": false  // 是否需要密码
  }
}
```

#### 下载证书

**接口**: `GET /api/cert`

**说明**: 下载 CA 证书文件

**响应**:
- Content-Type: `application/x-x509-ca-data`
- 文件名: `res-downloader-public.crt`

---

### 代理控制

#### 开启系统代理

**接口**: `POST /api/proxy/open`

**说明**: 设置系统代理为 127.0.0.1:8899

**请求**: 无需参数

**响应**:

```json
{
  "code": 1,
  "message": "ok",
  "data": {
    "value": true  // 代理状态
  }
}
```

**错误**:
- `401`: 需要 sudo 权限 (macOS/Linux)
- `500`: 设置失败

#### 关闭系统代理

**接口**: `POST /api/proxy/close`

**说明**: 清除系统代理设置

**响应**:

```json
{
  "code": 1,
  "message": "ok",
  "data": {
    "value": false
  }
}
```

#### 查询代理状态

**接口**: `GET /api/proxy/status`

**响应**:

```json
{
  "code": 1,
  "data": {
    "value": true
  }
}
```

---

### 配置管理

#### 获取配置

**接口**: `GET /api/config`

**响应**:

```json
{
  "code": 1,
  "data": {
    "Host": "127.0.0.1",
    "Port": "8899",
    "Theme": "lightTheme",
    "Locale": "zh",
    "Quality": 0,
    "SaveDirectory": "/Users/xxx/Downloads",
    "FilenameLen": 0,
    "FilenameTime": true,
    "UpstreamProxy": "",
    "OpenProxy": false,
    "DownloadProxy": false,
    "AutoProxy": false,
    "WxAction": true,
    "TaskNumber": 8,
    "DownNumber": 3,
    "UserAgent": "Mozilla/5.0...",
    "UseHeaders": "default",
    "InsertTail": true,
    "MimeMap": {
      "video/mp4": {"Type": "video", "Suffix": ".mp4"},
      "image/png": {"Type": "image", "Suffix": ".png"}
    },
    "Rule": "*"
  }
}
```

#### 更新配置

**接口**: `POST /api/config`

**请求**: 参考"获取配置"的响应结构

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

**注意事项**:
- 部分字段修改后需重启代理生效
- `MimeMap` 修改会影响资源类型识别

---

### 资源管理

#### 设置资源类型

**接口**: `POST /api/type`

**请求**:

```json
{
  "type": "video,audio,image"  // 逗号分隔
}
```

**说明**: 控制哪些类型的资源会被捕获

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

#### 清空资源列表

**接口**: `POST /api/clear`

**说明**: 清除内存中的所有已标记资源

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

#### 删除指定资源

**接口**: `POST /api/delete`

**请求**:

```json
{
  "sign": ["md5hash1", "md5hash2"]  // URL 的 MD5 签名列表
}
```

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

---

### 下载控制

#### 开始下载

**接口**: `POST /api/download`

**请求**:

```json
{
  "Id": "unique-id",
  "Url": "https://example.com/video.mp4",
  "UrlSign": "md5hash",
  "CoverUrl": "",
  "Domain": "example.com",
  "Classify": "video",
  "Suffix": ".mp4",
  "Status": "ready",
  "Size": 0,
  "ContentType": "video/mp4",
  "Description": "视频标题",
  "SavePath": "",
  "DecodeKey": "",
  "OtherData": {},
  "decodeStr": "base64-encoded-key"  // 可选,用于解密
}
```

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

**注意事项**:
- 下载是异步的,通过 WebSocket 事件推送进度
- `SavePath` 在下载完成后会被更新为实际路径

#### 取消下载

**接口**: `POST /api/cancel`

**请求**:

```json
{
  "Id": "unique-id"
}
```

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

#### 批量导出

**接口**: `POST /api/batch-export`

**请求**:

```json
{
  "content": "line1\nline2\nline3"  // 多行文本
}
```

**说明**: 将内容保存到文件并打开所在文件夹

**响应**:

```json
{
  "code": 1,
  "data": {
    "file_name": "/path/to/file.txt"
  }
}
```

---

### 系统交互

#### 打开文件夹

**接口**: `POST /api/open-folder`

**请求**:

```json
{
  "filePath": "/path/to/file"
}
```

**说明**: 在文件管理器中打开文件所在文件夹

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

#### 打开文件选择对话框

**接口**: `POST /api/open-file-dialog`

**请求**: 无需参数

**响应**:

```json
{
  "code": 1,
  "data": {
    "file": "/path/to/selected/file"
  }
}
```

#### 打开文件夹选择对话框

**接口**: `POST /api/open-directory-dialog`

**请求**: 无需参数

**响应**:

```json
{
  "code": 1,
  "data": {
    "folder": "/path/to/selected/folder"
  }
}
```

#### 设置系统密码

**接口**: `POST /api/set-system-password`

**请求**:

```json
{
  "password": "your-password",
  "isCache": true  // 是否缓存密码
}
```

**说明**: macOS/Linux 需要密码才能设置系统代理

**响应**:

```json
{
  "code": 1,
  "message": "ok"
}
```

---

### 应用信息

#### 获取应用信息

**接口**: `GET /api/app-info`

**响应**:

```json
{
  "code": 1,
  "data": {
    "AppName": "res-downloader",
    "Version": "3.1.3",
    "Description": "res-downloader是一款集网络资源...",
    "Copyright": "Copyright © 2023~2026",
    "IsProxy": true
  }
}
```

---

## Wails 绑定方法

Wails 绑定允许前端直接调用 Go 方法,无需 HTTP 请求。

### 导入

```typescript
import * as bind from "../../wailsjs/go/core/Bind"
```

### 可用方法

#### Config()

```typescript
const config = await bind.Config()
console.log(config.data)  // Config 对象
```

**返回**: `ResponseData` 对象,包含 `data` 字段为配置

#### AppInfo()

```typescript
const appInfo = await bind.AppInfo()
console.log(appInfo.data)  // App 对象
```

**返回**: `ResponseData` 对象,包含 `data` 字段为应用信息

#### ResetApp()

```typescript
await bind.ResetApp()
// 应用会重启
```

**说明**: 重置应用状态并重启(清除缓存、配置等)

---

## 前端事件

### 事件监听

前端通过 Wails `EventsOn` 监听后端推送的事件:

```typescript
import {EventsOn} from "../../wailsjs/runtime"

EventsOn("event", (res: string) => {
  const data = JSON.parse(res)
  const eventType = data.type
  const eventData = data.data

  // 分发事件
  handles[eventType]?.(eventData)
})
```

### 事件类型

#### newResources

**触发时机**: 捕获到新资源时

**数据结构**:

```json
{
  "type": "newResources",
  "data": {
    "Id": "unique-id",
    "Url": "https://example.com/video.mp4",
    "UrlSign": "md5hash",
    "CoverUrl": "https://example.com/cover.jpg",
    "Domain": "example.com",
    "Classify": "video",
    "Suffix": ".mp4",
    "Status": "ready",
    "Size": 1024000,
    "ContentType": "video/mp4",
    "Description": "视频标题",
    "SavePath": "",
    "DecodeKey": "",
    "OtherData": {}
  }
}
```

#### downloadProgress

**触发时机**: 下载进度更新时

**数据结构**:

```json
{
  "type": "downloadProgress",
  "data": {
    "Id": "unique-id",
    "Status": "running",
    "SavePath": "/path/to/file.mp4",
    "Message": "45%"
  }
}
```

**Status 枚举**:
- `ready`: 等待下载
- `pending`: 排队中
- `running`: 下载中
- `done`: 完成
- `error`: 失败

---

## 数据结构

### MediaInfo

```typescript
interface MediaInfo {
  Id: string              // 唯一标识
  Url: string             // 资源 URL
  UrlSign: string         // URL 的 MD5 签名
  CoverUrl: string        // 封面图 URL
  Domain: string          // 顶级域名
  Classify: string        // 资源类型: video/audio/image/m3u8/live
  Suffix: string          // 文件后缀: .mp4/.mp3/.jpg
  Status: string          // 状态: ready/pending/running/done/error
  Size: number            // 文件大小(字节)
  ContentType: string     // MIME 类型
  Description: string     // 描述信息(用于文件命名)
  SavePath: string        // 保存路径(下载完成后更新)
  DecodeKey: string       // 解密密钥(某些平台使用)
  OtherData: {            // 其他平台特定数据
    [key: string]: string
  }
}
```

### Config

```typescript
interface Config {
  Host: string            // 代理监听地址
  Port: string            // 代理端口
  Theme: string           // 主题: lightTheme/darkTheme
  Locale: string          // 语言: zh/en
  Quality: number         // 视频质量: 0/1/2
  SaveDirectory: string   // 下载保存目录
  FilenameLen: number     // 文件名长度限制(0=不限制)
  FilenameTime: boolean   // 是否添加时间戳
  UpstreamProxy: string   // 上游代理地址
  OpenProxy: boolean      // 是否使用上游代理(请求拦截)
  DownloadProxy: boolean  // 是否使用上游代理(下载)
  AutoProxy: boolean      // 是否自动开启代理
  WxAction: boolean       // 微信视频号抓取模式
  TaskNumber: number      // 下载线程数
  DownNumber: number      // 并发下载数
  UserAgent: string       // User-Agent
  UseHeaders: string      // 使用的 Headers: default/all/custom
  InsertTail: boolean     // 是否插入到列表尾部
  MimeMap: {              // MIME 类型映射
    [mimeType: string]: {
      Type: string        // 资源类型
      Suffix: string      // 文件后缀
    }
  }
  Rule: string            // 域名匹配规则
}
```

### ResponseData

```typescript
interface ResponseData {
  code: number            // 1=成功, 0=失败
  message: string         // 消息
  data?: any             // 数据
}
```

---

## 错误码

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权(需要密码) |
| 404 | 接口不存在 |
| 500 | 服务器内部错误 |

### 业务错误码

| code | message | 说明 |
|------|---------|------|
| 0 | 错误消息 | 通用错误 |
| 1 | ok | 成功 |

### 常见错误场景

1. **证书安装失败**
```json
{
  "code": 0,
  "message": "Access is denied",
  "data": {
    "isPass": true
  }
}
```

2. **代理设置失败**
```json
{
  "code": 0,
  "message": "permission denied",
  "data": {
    "value": false
  }
}
```

3. **下载目录不存在**
```json
{
  "code": 0,
  "message": "save directory not set"
}
```

---

## 使用示例

### 完整的下载流程

```typescript
import appApi from "@/api/app"

// 1. 开始下载
const startDownload = async (media: MediaInfo) => {
  const res = await appApi.download({
    ...media,
    decodeStr: ''  // 可选的解密密钥
  })

  if (res.code === 0) {
    console.error('下载失败:', res.message)
  }
}

// 2. 监听进度
eventStore.addHandle({
  type: "downloadProgress",
  event: (progress) => {
    console.log(`${progress.Id}: ${progress.Message}`)

    if (progress.Status === 'done') {
      console.log('下载完成:', progress.SavePath)
    }
  }
})

// 3. 取消下载
const cancelDownload = async (id: string) => {
  const res = await appApi.cancel({id})
  if (res.code === 1) {
    console.log('已取消')
  }
}
```

### 配置更新流程

```typescript
// 1. 获取当前配置
const configRes = await appApi.getConfig()
const config = configRes.data

// 2. 修改配置
config.SaveDirectory = '/new/path'
config.TaskNumber = 16

// 3. 提交更新
const updateRes = await appApi.setConfig(config)
if (updateRes.code === 1) {
  console.log('配置已更新')
}
```

### 资源导出流程

```typescript
const exportResources = (items: MediaInfo[]) => {
  // 导出为 JSON
  const content = items
    .map(item => encodeURIComponent(JSON.stringify(item)))
    .join('\n')

  appApi.batchExport({content})
    .then(res => {
      if (res.code === 1) {
        alert(`已导出到: ${res.data.file_name}`)
      }
    })
}
```

---

**最后更新**: 2026-01-28
