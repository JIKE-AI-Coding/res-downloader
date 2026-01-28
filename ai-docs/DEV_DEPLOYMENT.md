# 部署与发布指南

> 本文档介绍 res-downloader 的编译、打包和发布流程。

---

## 目录

- [环境准备](#环境准备)
- [开发模式](#开发模式)
- [生产构建](#生产构建)
- [跨平台编译](#跨平台编译)
- [签名与公证](#签名与公证)
- [自动发布](#自动发布)

---

## 环境准备

### 系统要求

| 平台 | 最低要求 | 推荐配置 |
|------|----------|----------|
| **Windows** | Windows 10 1903+ | Windows 11 |
| **macOS** | macOS 10.15+ | macOS 13+ |
| **Linux** | glibc 2.17+ | Ubuntu 20.04+ |

### 必需工具

```bash
# 1. Go (1.18+)
go version

# 2. Node.js (16+)
node -v
npm -v

# 3. Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails version
```

### 平台特定依赖

#### Windows

```bash
# 安装 MSVC (Visual Studio Build Tools)
# 下载: https://visualstudio.microsoft.com/downloads/

# 或使用 mingw-w64
# 推荐使用 TDM-GCC: https://jmeubank.github.io/tdm-gcc/
```

#### macOS

```bash
# 安装 Xcode Command Line Tools
xcode-select --install

# 如果开发 iOS 应用,安装完整 Xcode
```

#### Linux

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gcc gcc-c++ make gtk3-devel webkit2gtk3-devel

# Arch Linux
sudo pacman -S base-devel gtk3 webkit2gtk
```

---

## 开发模式

### 启动开发服务器

```bash
# 克隆项目
git clone https://github.com/putyy/res-downloader.git
cd res-downloader

# 安装前端依赖
cd frontend && npm install && cd ..

# 启动开发模式(热重载)
wails dev
```

**特性**:
- 前端代码修改后自动刷新
- Go 代码修改后自动重启
- 日志直接输出到终端
- 支持断点调试

### 调试配置

#### VS Code

创建 `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Wails App",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}",
      "env": {
        "WAILS_DEV": "1"
      },
      "args": ["dev"],
      "showLog": true
    }
  ]
}
```

---

## 生产构建

### 单平台构建

```bash
# 构建当前平台的可执行文件
wails build

# 指定输出目录
wails build -o ./output/res-downloader

# 启用压缩
wails build -clean

# 跳过前端构建(使用已有资源)
wails build -skipfrontend
```

**输出位置**:
- Windows: `build/bin/res-downloader.exe`
- macOS: `build/bin/res-downloader.app`
- Linux: `build/bin/res-downloader`

### 构建配置

编辑 `wails.json`:

```json
{
  "name": "res-downloader",
  "outputfilename": "res-downloader",
  "author": {
    "name": "putyy",
    "email": "putyy@qq.com"
  },
  "info": {
    "companyName": "res-downloader",
    "productName": "res-downloader",
    "productVersion": "3.1.3",
    "copyright": "Copyright © 2023"
  }
}
```

---

## 跨平台编译

### 使用 Docker

**优点**: 环境隔离,无需手动配置各平台依赖

#### 1. 创建 Dockerfile

```dockerfile
# Dockerfile.linux
FROM golang:1.21-bullseye AS build

RUN apt-get update && apt-get install -y \
    build-essential libgtk-3-dev libwebkit2gtk-4.0-dev \
    npm && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN cd frontend && npm install && npm run build
ENV CGO_ENABLED=1
RUN wails build -platform linux/amd64

FROM debian:bullseye-slim
WORKDIR /root/
COPY --from=build /app/build/bin/ /usr/local/bin/
CMD ["res-downloader"]
```

#### 2. 构建镜像

```bash
# Linux AMD64
docker build -f Dockerfile.linux -t res-downloader:linux-amd64 .
docker run --rm -v $(pwd)/dist:/dist res-downloader:linux-amd64 \
  cp /usr/local/bin/res-downloader /dist/

# Linux ARM64
docker build --platform linux/arm64 -f Dockerfile.linux \
  -t res-downloader:linux-arm64 .
```

### 本地交叉编译

#### Windows → Linux

```bash
# 1. 安装交叉编译工具
# 下载 MinGW-w64: https://www.mingw-w64.org/

# 2. 设置环境变量
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="1"

# 3. 编译
wails build -platform linux/amd64
```

#### macOS → Windows

```bash
# 1. 安装 MinGW
brew install mingw-w64

# 2. 编译
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++ \
    wails build -platform windows/amd64
```

### GitHub Actions

创建 `.github/workflows/build.yml`:

```yaml
name: Build

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        include:
          - goos: windows
            goarch: amd64
            artifact: res-downloader-windows-amd64.exe
          - goos: darwin
            goarch: amd64
            artifact: res-downloader-darwin-amd64.zip
          - goos: darwin
            goarch: arm64
            artifact: res-downloader-darwin-arm64.zip
          - goos: linux
            goarch: amd64
            artifact: res-downloader-linux-amd64

    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: 18

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build frontend
        run: |
          cd frontend
          npm install
          npm run build

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 1
        run: |
          if [ "$GOOS" = "windows" ]; then
            sudo apt-get install mingw-w64
            export CC=x86_64-w64-mingw32-gcc
            export CXX=x86_64-w64-mingw32-g++
          elif [ "$GOOS" = "darwin" ]; then
            export CC=o64-clang
            export CXX=o64-clang++
          fi
          wails build -platform $GOOS/$GOARCH

      - name: Package
        run: |
          cd build/bin
          if [ "$GOOS" = "darwin" ]; then
            zip -r ../../../${{ matrix.artifact }} *
          else
            mv res-downloader* ../../../${{ matrix.artifact }}
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: ${{ matrix.artifact }}
          path: ${{ matrix.artifact }}
```

---

## 签名与公证

### Windows 代码签名

#### 1. 获取证书

- 从 CA 机构购买代码签名证书
- 导出为 `.pfx` 文件
- 记录证书密码

#### 2. 配置签名

```bash
# 方法 1: 环境变量
export CSC_LINK=path/to/certificate.pfx
export CSC_KEY_PASSWORD=your-password

# 方法 2: Wails 参数
wails build \
  -sign \
  -certkeypath=path/to/certificate.pfx \
  -certkeypassword=your-password
```

#### 3. 验证签名

```bash
# Windows
signtool verify /pa /v res-downloader.exe

# 查看签名信息
signtool verify /pa res-downloader.exe
```

### macOS 代码签名

#### 1. 创建开发者证书

1. 登录 Apple Developer
2. 创建 "Developer ID Application" 证书
3. 下载并安装到钥匙串

#### 2. 配置签名

```bash
# 方法 1: 自动签名(推荐)
wails build -sign

# 方法 2: 指定证书
wails build \
  -sign \
  -certkeypath="Developer ID Application: Your Name (TEAM_ID)"

# 方法 3: 硬编码签名(不推荐)
# 在 main.go 中设置:
// #include <Security/Security.h>
// extern const uint8_t g_cert_data[] = { ... };
// extern const uint8_t g_cert_key[] = { ... };
```

#### 3. 公证(Notarization)

**必需**: macOS 10.15+ 需要公证才能运行

```bash
# 1. 生成公证文件
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: Your Name" \
  --options runtime \
  res-downloader.app

# 2. 提交公证
xcrun notarytool submit res-downloader.app.zip \
  --apple-id "your@email.com" \
  --password "app-specific-password" \
  --team-id "TEAM_ID" \
  --wait

# 3. 装订公证票据
xcrun stapler staple res-downloader.app

# 4. 验证
xcrun stapler validate res-downloader.app
spctl -a -vvv -t execute res-downloader.app
```

#### 4. 自动化公证

创建 `notarize.sh`:

```bash
#!/bin/bash

APPLE_ID="your@email.com"
PASSWORD="app-specific-password"
TEAM_ID="TEAM_ID"
APP_PATH="build/bin/res-downloader.app"

# 签名
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: Your Name ($TEAM_ID)" \
  --options runtime \
  "$APP_PATH"

# 打包
zip -r res-downloader.app.zip "$APP_PATH"

# 公证
xcrun notarytool submit res-downloader.app.zip \
  --apple-id "$APPLE_ID" \
  --password "$PASSWORD" \
  --team-id "$TEAM_ID" \
  --wait

# 装订
xcrun stapler staple "$APP_PATH"

# 验证
xcrun stapler validate "$APP_PATH"
spctl -a -vvv -t execute "$APP_PATH"

echo "Notarization complete!"
```

---

## 自动发布

### GitHub Actions 完整流程

创建 `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    strategy:
      matrix:
        goos: [windows, darwin, linux]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
          - goos: darwin
            goarch: amd64  # 如果只支持 ARM64

    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: 18

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y \
            build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
          npm install -g @wailsapp/cli

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 1
        run: |
          if [ "$GOOS" = "windows" ]; then
            sudo apt-get install -y mingw-w64
            export CC=x86_64-w64-mingw32-gcc
            export CXX=x86_64-w64-mingw32-g++
          elif [ "$GOOS" = "darwin" ]; then
            export CC=o64-clang
            export CXX=o64-clang++
          fi
          wails build -platform $GOOS/$GOARCH -clean

      - name: Package
        run: |
          cd build/bin
          if [ "$GOOS" = "darwin" ]; then
            zip -r ../../../res-downloader-${{ matrix.goos }}-${{ matrix.goarch }}.zip *
          elif [ "$GOOS" = "windows" ]; then
            mv res-downloader.exe ../../../res-downloader-${{ matrix.goos }}-${{ matrix.goarch }}.exe
          else
            tar czf ../../../res-downloader-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz *
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: res-downloader-${{ matrix.goos }}-${{ matrix.goarch }}
          path: res-downloader-*

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: res-downloader-*
          draft: false
          prerelease: false
```

### 版本号管理

#### 1. 语义化版本

遵循 [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH` (例如: 3.1.3)
- MAJOR: 不兼容的 API 变更
- MINOR: 向后兼容的功能新增
- PATCH: 向后兼容的问题修复

#### 2. 自动版本号

```bash
# 方法 1: Git tags
git tag v3.1.3
git push origin v3.1.3

# 方法 2: 从代码生成
VERSION=$(grep "productVersion" wails.json | sed 's/.*"\(.*\)".*/\1/')
echo "Current version: $VERSION"
```

---

## 性能优化

### 减小体积

```bash
# 1. 启用 UPX 压缩
upx --best --lzma res-downloader

# 2. 剥离调试符号
go build -ldflags="-s -w" .

# 3. 前端资源优化
cd frontend
npm run build  # 已启用压缩
```

### 启动速度优化

```go
// main.go
func main() {
    // 延迟加载非关键模块
    app := core.GetApp(assets, wailsJson)

    // 使用 lazy initialization
    // ...
}
```

---

## 发布检查清单

### 构建前

- [ ] 更新 `wails.json` 中的版本号
- [ ] 更新 `README.md` 中的版本信息
- [ ] 运行所有测试
- [ ] 检查依赖安全性
- [ ] 更新 CHANGELOG

### 构建时

- [ ] 跨平台构建成功
- [ ] 代码签名通过
- [ ] 公证通过(macOS)
- [ ] 文件大小合理(< 100MB)

### 发布后

- [ ] 创建 GitHub Release
- [ ] 上传到其他镜像站(可选)
- [ ] 更新官网
- [ ] 发布公告

---

**最后更新**: 2026-01-28
