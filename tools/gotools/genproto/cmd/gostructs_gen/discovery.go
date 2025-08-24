package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// 添加 Google protobuf 标准库路径
// 动态检测protobuf include路径
func findDescriptorProto() (protobufIncludePath string) {
	pbIncludePaths := getProtobufIncludePaths()

	protobufPathFound := false
	for _, path := range pbIncludePaths {
		descriptorPath := filepath.Join(path, "google/protobuf/descriptor.proto")

		if info, err := os.Stat(descriptorPath); err == nil && !info.IsDir() {
			fmt.Printf("Found protobuf well-known types at: %s\n", path)
			protobufIncludePath = path
			protobufPathFound = true
			break
		}
	}

	if !protobufPathFound {
		fmt.Printf("Warning: protobuf well-known types not found in detected paths: %v\n", pbIncludePaths)
	}

	return
}

// getProtobufIncludePaths 动态获取protobuf include路径
func getProtobufIncludePaths() []string {
	var paths []string

	// 1. 检查专用环境变量
	if protobufInclude := os.Getenv("PROTOBUF_INCLUDE_PATH"); protobufInclude != "" {
		paths = append(paths, protobufInclude)
	}

	// 2. 基于 protoc 位置推断路径 (跨平台)
	if protocPath, err := exec.LookPath("protoc"); err == nil {
		// protoc在 /opt/homebrew/bin/protoc -> include在 /opt/homebrew/include
		// 或者 C:\tools\protoc\bin\protoc.exe -> C:\tools\protoc\include
		if binDir := filepath.Dir(protocPath); binDir != "" {
			if rootDir := filepath.Dir(binDir); rootDir != "" {
				paths = append(paths, filepath.Join(rootDir, "include"))
			}
		}
	}

	// 3. 添加操作系统特定路径
	paths = append(paths, getOSSpecificPaths()...)

	// 去重
	seen := make(map[string]bool)
	var uniquePaths []string
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	return uniquePaths
}

// getOSSpecificPaths 根据操作系统返回特定的protobuf路径
func getOSSpecificPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return getWindowsPaths()
	case "darwin":
		return getMacOSPaths()
	case "linux":
		return getLinuxPaths()
	default:
		// 对于其他系统，返回通用路径
		return []string{"/usr/local/include", "/usr/include"}
	}
}

// getWindowsPaths 获取Windows系统的protobuf路径
func getWindowsPaths() []string {
	var paths []string

	// 1. vcpkg 安装路径 (最常见)
	if vcpkgRoot := os.Getenv("VCPKG_ROOT"); vcpkgRoot != "" {
		triplets := []string{"x64-windows", "x86-windows", "x64-windows-static", "x86-windows-static"}
		for _, triplet := range triplets {
			paths = append(paths, filepath.Join(vcpkgRoot, "installed", triplet, "include"))
		}
	}

	// 常见的vcpkg默认路径
	commonVcpkgPaths := []string{
		"C:\\vcpkg\\installed\\x64-windows\\include",
		"C:\\vcpkg\\installed\\x86-windows\\include",
		"C:\\tools\\vcpkg\\installed\\x64-windows\\include",
	}
	paths = append(paths, commonVcpkgPaths...)

	// 2. Chocolatey 安装路径
	if chocoInstall := os.Getenv("ChocolateyInstall"); chocoInstall != "" {
		paths = append(paths, filepath.Join(chocoInstall, "lib", "protoc", "tools", "include"))
	}
	// Chocolatey 默认路径
	paths = append(paths, "C:\\ProgramData\\chocolatey\\lib\\protoc\\tools\\include")

	// 3. MSYS2/MinGW 路径
	msys2Paths := []string{
		"C:\\msys64\\mingw64\\include",
		"C:\\msys64\\mingw32\\include",
		"C:\\msys64\\usr\\include",
	}
	paths = append(paths, msys2Paths...)

	// 4. 常见手动安装路径
	commonPaths := []string{
		"C:\\Program Files\\protobuf\\include",
		"C:\\Program Files (x86)\\protobuf\\include",
		"C:\\protobuf\\include",
		"C:\\tools\\protobuf\\include",
	}
	paths = append(paths, commonPaths...)

	// 5. 基于环境变量PATH中的protoc.exe推断
	if pathEnv := os.Getenv("PATH"); pathEnv != "" {
		pathDirs := strings.Split(pathEnv, ";")
		for _, dir := range pathDirs {
			if strings.Contains(strings.ToLower(dir), "protoc") || strings.Contains(strings.ToLower(dir), "protobuf") {
				// 尝试从bin目录推断include目录
				if strings.HasSuffix(strings.ToLower(dir), "bin") {
					parentDir := filepath.Dir(dir)
					paths = append(paths, filepath.Join(parentDir, "include"))
				}
			}
		}
	}

	return paths
}

// getMacOSPaths 获取macOS系统的protobuf路径
func getMacOSPaths() []string {
	var paths []string

	// Homebrew 支持
	if homebrewPrefix := os.Getenv("HOMEBREW_PREFIX"); homebrewPrefix != "" {
		paths = append(paths, filepath.Join(homebrewPrefix, "include"))
	}

	// 常见 macOS 路径
	fallbackPaths := []string{
		"/opt/homebrew/include", // Homebrew on Apple Silicon
		"/usr/local/include",    // Homebrew on Intel Mac
		"/opt/local/include",    // MacPorts
	}
	paths = append(paths, fallbackPaths...)

	return paths
}

// getLinuxPaths 获取Linux系统的protobuf路径
func getLinuxPaths() []string {
	var paths []string

	// 检查常见的包管理器路径
	linuxPaths := []string{
		"/usr/include",                   // 系统安装
		"/usr/local/include",             // 手动编译安装
		"/opt/protobuf/include",          // 可选安装位置
		"/snap/protobuf/current/include", // Snap包
	}
	paths = append(paths, linuxPaths...)

	// 检查是否使用Linuxbrew
	if homebrewPrefix := os.Getenv("HOMEBREW_PREFIX"); homebrewPrefix != "" {
		paths = append(paths, filepath.Join(homebrewPrefix, "include"))
	}

	return paths
}
