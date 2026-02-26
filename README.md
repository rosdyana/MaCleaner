# 🧹 MaCleaner

A lightweight, fast, and efficient macOS storage cleaner built with Go. No heavy dependencies, no bloated UI - just clean.

[![Build Status](https://github.com/rosdyana/MaCleaner/actions/workflows/release.yml/badge.svg)](https://github.com/rosdyana/MaCleaner/actions)
[![Release](https://img.shields.io/github/release/rosdyana/MaCleaner.svg)](https://github.com/rosdyana/MaCleaner/releases/latest)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)

![Size](https://img.shields.io/badge/size-2.0MB-brightgreen.svg)
![Dependencies](https://img.shields.io/badge/dependencies-1-blue.svg)

## ✨ Features

- **🧽 Storage Cleanup** - Clean caches, logs, temp files, and more
- **📦 Big Files Finder** - Find and remove large files taking up space
- **🔁 Duplicate Finder** - Find and delete duplicate files
- **📅 Old Files Finder** - Find files not accessed in 30/90/180/365 days
- **⚡ Fast & Lightweight** - Only 2MB binary, minimal dependencies
- **🎨 Terminal UI** - Simple and intuitive text-based interface
- **🔒 Safe** - Shows what will be deleted before cleaning

## 🚀 Quick Start

### Download Pre-built Binary

Download the latest release from [Releases page](https://github.com/rosdyana/MaCleaner/releases/latest)

```bash
# For Apple Silicon (M1/M2/M3)
curl -L -o macos-cleaner https://github.com/rosdyana/MaCleaner/releases/latest/download/macos-cleaner-arm64
chmod +x macos-cleaner

# For Intel Macs
curl -L -o macos-cleaner https://github.com/rosdyana/MaCleaner/releases/latest/download/macos-cleaner-amd64
chmod +x macos-cleaner
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/rosdyana/MaCleaner.git
cd MaCleaner

# Build
./build.sh

# Or manually
go build -trimpath -ldflags="-s -w -buildid=" -o bin/macos-cleaner .
```

## 📖 Usage

```bash
./macos-cleaner
```

### Main Menu

```
🧹 macOS Storage Cleaner

Choose an option:

[1] 🧽 Storage Cleanup - Clean caches, logs, temp files
[2] 📦 Big Files Finder - Find large files taking up space
[3] 🔁 Duplicate Finder - Find duplicate files
[4] 📅 Old Files Finder - Find files not accessed recently

Press 1-4 to select, q to quit
```

### Storage Cleanup

Navigate with arrow keys, select with `Space`, then press `s` to scan:

```
Cache:
  [✓] User Caches               Application caches
  [✓] Safari Cache              Safari browser cache
  [✓] Chrome Cache              Chrome browser cache

Logs:
  [ ] User Logs                 Application logs
  [ ] Crash Reports             App crash logs

Package Manager:
  [✓] Homebrew Cache            Homebrew download cache
  [✓] npm Cache                 npm packages cache

[↑↓] Navigate  [Space] Toggle  [a] All  [n] None  [s] Scan  [b] Back  [q] Quit
```

### Big Files Finder

Find files larger than 100MB/500MB/1GB/5GB:

```
🧹 Big Files Results (> 500 MB)

Found 12 large files:

  [ ]  1.2 GB    45d  .../Downloads/movie.mkv
> [✓]  3.5 GB    120d .../Documents/backup.dmg
  [ ]  850 MB    12d  .../Downloads/installer.pkg

Selected: 1 files (3.5 GB)

[↑↓] Navigate  [Space] Toggle  [a] All  [d] Delete  [b] Back  [q] Quit
```

### Duplicate Finder

```
🧹 Duplicate Files Results

Found 5 duplicate groups:

  [ ] Group 1: 15 MB (3 files)
    ├─ .../Documents/photo1.jpg
    ├─ .../Desktop/photo1.jpg
    └─ .../Downloads/photo1.jpg

> [✓] Group 2: 250 MB (2 files)
    └─ .../Videos/movie.mp4

Selected: 1 groups (saves 250 MB)

[↑↓] Navigate  [Space] Toggle  [d] Delete Selected  [b] Back  [q] Quit
```

## 🎯 Cleanup Targets

### Cache Files
- User Caches (`~/Library/Caches`)
- Safari, Chrome, Firefox caches
- App Store, iCloud, Photos caches
- Quick Look thumbnails

### Log Files
- User Logs (`~/Library/Logs`)
- System Logs (`/var/log`)
- Crash Reports
- Diagnostic Logs

### Development Files
- Xcode Derived Data
- iOS Simulator files
- Android Build Cache
- Gradle Cache

### Package Manager Caches
- Homebrew (`brew cleanup`)
- npm, yarn, pip
- Cargo, Composer, gem
- CocoaPods

### App Caches
- Spotify, Slack, Discord
- Teams, Zoom, VS Code

## 🛠️ Development

### Project Structure

```
MaCleaner/
├── bin/                    # Build output
├── internal/
│   ├── cleaner/           # File deletion logic
│   ├── ltui/              # Lightweight terminal UI
│   ├── models/            # Data types & cleanup targets
│   ├── scanner/           # File scanning logic
│   └── utils/             # Path, sudo utilities
├── build.sh               # Build script
├── go.mod
├── main.go
└── README.md
```

### Running Tests

```bash
go test ./internal/... -v
```

### Build Options

```bash
# Current architecture only
go build -o macos-cleaner .

# Intel Macs
GOOS=darwin GOARCH=amd64 go build -o macos-cleaner-amd64 .

# Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o macos-cleaner-arm64 .

# Universal binary
lipo -create -output macos-cleaner-universal macos-cleaner-amd64 macos-cleaner-arm64
```

## ⚙️ Technical Details

### Size Comparison

| Version | Size | Reduction |
|---------|------|-----------|
| Original (Bubble Tea) | 4.6 MB | - |
| **Lightweight TUI** | **2.0 MB** | **-56%** |

### Dependencies

- Only 1 external dependency: `golang.org/x/sys`
- No Bubble Tea, Lipgloss, or other heavy TUI frameworks
- Pure Go implementation with minimal ANSI escape codes

### Performance

- Scans `~/Documents`, `~/Desktop`, `~/Downloads`, `~/Movies`, `~/Music`, `~/Pictures`
- Skips `~/Library` and hidden directories
- Progress updates every 500 files
- Multi-threaded file operations

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org)
- Inspired by various macOS cleanup tools
- Thanks to all contributors

## ⚠️ Disclaimer

This tool modifies files on your system. Always review what will be deleted before confirming cleanup. The authors are not responsible for any data loss.

---

Made with ❤️ by [Rosdyana](https://github.com/rosdyana)
