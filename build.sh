#!/bin/bash
# build.sh - 构建 go-stress-test 可执行文件

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="go-stress-test"
BUILD_DIR="$SCRIPT_DIR/cmd/go-stress-test"

echo "🚀 Building $BINARY_NAME from $BUILD_DIR..."

# 编译
go build -o "$SCRIPT_DIR/$BINARY_NAME" "$BUILD_DIR"

# 赋予执行权限（Linux/macOS）
chmod +x "$SCRIPT_DIR/$BINARY_NAME"

echo "Build successful! Executable created: $SCRIPT_DIR/$BINARY_NAME"
echo "Run it with: ./$BINARY_NAME --help"