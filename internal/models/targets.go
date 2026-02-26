// Package models provides the default cleanup targets configuration
package models

// GetDefaultTargets returns the default list of cleanup targets
func GetDefaultTargets() []CleanupTarget {
	return []CleanupTarget{
		// ===== CACHE FILES =====
		{Name: "User Caches", Path: "~/Library/Caches/*", Description: "Application caches", Category: "Cache", RequiresSudo: false},
		{Name: "System Caches", Path: "/Library/Caches/*", Description: "System-wide caches", Category: "Cache", RequiresSudo: true},
		{Name: "Safari Cache", Path: "~/Library/Caches/com.apple.Safari/*", Description: "Safari browser cache", Category: "Cache", RequiresSudo: false},
		{Name: "Chrome Cache", Path: "~/Library/Caches/Google/Chrome/*/Cache/*", Description: "Chrome browser cache", Category: "Cache", RequiresSudo: false},
		{Name: "Firefox Cache", Path: "~/Library/Caches/Firefox/Profiles/*/cache2/*", Description: "Firefox browser cache", Category: "Cache", RequiresSudo: false},
		{Name: "Quick Look Cache", Path: "/private/var/folders/*/C/com.apple.QuickLook.thumbnailcache/*", Description: "Quick Look thumbnails", Category: "Cache", RequiresSudo: false},
		{Name: "iCloud Cache", Path: "~/Library/Caches/CloudKit/*", Description: "iCloud sync cache", Category: "Cache", RequiresSudo: false},
		{Name: "Photos Cache", Path: "~/Library/Containers/com.apple.Photos/Data/Library/Caches/*", Description: "Photos app cache", Category: "Cache", RequiresSudo: false},
		{Name: "App Store Cache", Path: "~/Library/Caches/com.apple.appstore/*", Description: "App Store cache", Category: "Cache", RequiresSudo: false},

		// ===== LOG FILES =====
		{Name: "User Logs", Path: "~/Library/Logs/*", Description: "Application logs", Category: "Logs", RequiresSudo: false},
		{Name: "System Logs", Path: "/var/log/*", Description: "System log files", Category: "Logs", RequiresSudo: true},
		{Name: "Crash Reports", Path: "~/Library/Application Support/CrashReporter/*", Description: "App crash logs", Category: "Logs", RequiresSudo: false},
		{Name: "Diagnostic Logs", Path: "/private/var/db/diagnostics/*", Description: "System diagnostics", Category: "Logs", RequiresSudo: true},

		// ===== TEMP FILES =====
		{Name: "User Temp", Path: "/private/var/tmp/*", Description: "User temporary files", Category: "Temp", RequiresSudo: true},
		{Name: "System Temp", Path: "/private/tmp/*", Description: "System temporary files", Category: "Temp", RequiresSudo: true},
		{Name: "Var Folders", Path: "/var/folders/*/*/T/*", Description: "System temp folders", Category: "Temp", RequiresSudo: false},

		// ===== TRASH =====
		{Name: "Trash", Path: "~/.Trash/*", Description: "Files in Trash", Category: "Trash", RequiresSudo: false},

		// ===== XCODE / DEVELOPMENT =====
		{Name: "Xcode Derived Data", Path: "~/Library/Developer/Xcode/DerivedData/*", Description: "Xcode build artifacts", Category: "Dev", RequiresSudo: false},
		{Name: "Xcode Archives", Path: "~/Library/Developer/Xcode/Archives/*", Description: "Xcode archives", Category: "Dev", RequiresSudo: false},
		{Name: "Xcode Device Support", Path: "~/Library/Developer/Xcode/iOS DeviceSupport/*", Description: "iOS debugging symbols", Category: "Dev", RequiresSudo: false},
		{Name: "iOS Simulator", Path: "~/Library/Developer/CoreSimulator/*", Description: "iOS Simulator files", Category: "Dev", RequiresSudo: false},
		{Name: "Android Build Cache", Path: "~/.android/build-cache", Description: "Android build cache", Category: "Dev", RequiresSudo: false},
		{Name: "Gradle Cache", Path: "~/.gradle/caches", Description: "Gradle build cache", Category: "Dev", RequiresSudo: false},

		// ===== PACKAGE MANAGERS =====
		{Name: "Homebrew Cache", Path: "~/Library/Caches/Homebrew", Description: "Homebrew download cache", Category: "Package Manager", RequiresSudo: false, IsCommand: true, Command: "brew cleanup"},
		{Name: "npm Cache", Path: "~/.npm/*", Description: "npm packages cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "yarn Cache", Path: "~/Library/Caches/yarn/*", Description: "yarn packages cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "Cargo Cache", Path: "~/.cargo/registry/cache/*", Description: "Rust crates cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "Cargo Git", Path: "~/.cargo/git/checkouts/*", Description: "Cargo git checkouts", Category: "Package Manager", RequiresSudo: false},
		{Name: "pip Cache", Path: "~/Library/Caches/pip/*", Description: "Python pip cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "Composer Cache", Path: "~/Library/Caches/composer/*", Description: "PHP Composer cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "gem Cache", Path: "~/.gem/cache/*", Description: "Ruby gems cache", Category: "Package Manager", RequiresSudo: false},
		{Name: "CocoaPods Cache", Path: "~/Library/Caches/CocoaPods/*", Description: "CocoaPods cache", Category: "Package Manager", RequiresSudo: false},

		// ===== APP CACHES =====
		{Name: "Spotify Cache", Path: "~/Library/Caches/com.spotify.client/*", Description: "Spotify offline cache", Category: "Apps", RequiresSudo: false},
		{Name: "Slack Cache", Path: "~/Library/Containers/com.tinyspeck.slackmacgap/Data/Library/Application Support/Slack/Cache/*", Description: "Slack cache", Category: "Apps", RequiresSudo: false},
		{Name: "Discord Cache", Path: "~/Library/Application Support/discord/Cache/*", Description: "Discord cache", Category: "Apps", RequiresSudo: false},
		{Name: "Teams Cache", Path: "~/Library/Application Support/Microsoft/Teams/*", Description: "Microsoft Teams cache", Category: "Apps", RequiresSudo: false},
		{Name: "Zoom Cache", Path: "~/Library/Caches/us.zoom.xos/*", Description: "Zoom cache", Category: "Apps", RequiresSudo: false},
		{Name: "VS Code Cache", Path: "~/Library/Application Support/Code/Cache/*", Description: "VS Code cache", Category: "Apps", RequiresSudo: false},

		// ===== SYSTEM / HIDDEN =====
		{Name: "Saved App State", Path: "~/Library/Saved Application State/*", Description: "App state data", Category: "System", RequiresSudo: false},
		{Name: "Mail Downloads", Path: "~/Library/Containers/com.apple.mail/Data/Library/Mail Downloads/*", Description: "Mail attachments", Category: "System", RequiresSudo: false},
		{Name: "Message Attachments", Path: "~/Library/Messages/Attachments/*", Description: "iMessage photos/videos", Category: "System", RequiresSudo: false},
		{Name: "QuickTime Cache", Path: "~/Library/Caches/com.apple.QuickTime*", Description: "QuickTime cache", Category: "System", RequiresSudo: false},

		// ===== BACKUPS =====
		{Name: "iOS Backups", Path: "~/Library/Application Support/MobileSync/Backup/*", Description: "iPhone/iPad backups", Category: "Backups", RequiresSudo: false},
		{Name: "Time Machine Local", Path: "", Description: "Time Machine local snapshots", Category: "Backups", RequiresSudo: true, IsCommand: true, Command: "tmutil deletelocalsnapshots /"},

		// ===== DOWNLOADS (Optional) =====
		{Name: "Downloads", Path: "~/Downloads/*", Description: "Downloads folder", Category: "User", RequiresSudo: false},
	}
}

// HasSelection checks if any target is selected
func HasSelection(targets []CleanupTarget) bool {
	for _, t := range targets {
		if t.Selected {
			return true
		}
	}
	return false
}

// HasBigFilesSelection checks if any big file is selected
func HasBigFilesSelection(selected map[int]bool) bool {
	for _, v := range selected {
		if v {
			return true
		}
	}
	return false
}

// HasDuplicateSelection checks if any duplicate group is selected
func HasDuplicateSelection(selected map[int]bool) bool {
	for _, v := range selected {
		if v {
			return true
		}
	}
	return false
}

// HasOldFilesSelection checks if any old file is selected
func HasOldFilesSelection(selected map[int]bool) bool {
	for _, v := range selected {
		if v {
			return true
		}
	}
	return false
}
