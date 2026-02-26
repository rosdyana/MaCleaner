package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"macos-cleaner/internal/cleaner"
	"macos-cleaner/internal/ltui"
	"macos-cleaner/internal/models"
	"macos-cleaner/internal/scanner"
	"macos-cleaner/internal/utils"
)

type app struct {
	term         *ltui.Terminal
	scanner      *scanner.Scanner
	cleaner      *cleaner.Cleaner
	targets      []models.CleanupTarget
	bigFiles     []models.BigFile
	duplicateGroups []models.DuplicateGroup
	oldFiles     []models.OldFile
	
	// State
	cursor       int
	selections   map[int]bool
}

func newApp() *app {
	sudoMgr := utils.NewSudoManager()
	return &app{
		term:       ltui.NewTerminal(),
		scanner:    scanner.New(sudoMgr),
		cleaner:    cleaner.New(sudoMgr),
		targets:    models.GetDefaultTargets(),
		selections: make(map[int]bool),
	}
}

func (a *app) run() {
	defer a.term.ShowCursor()
	a.term.HideCursor()

	for {
		key := a.term.PrintMenu()
		switch key {
		case "1":
			a.runCleanup()
		case "2":
			a.runBigFiles()
		case "3":
			a.runDuplicates()
		case "4":
			a.runOldFiles()
		case "q", "Q":
			return
		}
	}
}

func (a *app) runCleanup() {
	a.cursor = 0
	for {
		key := a.term.PrintTargets(a.targets, a.cursor)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			if a.cursor < len(a.targets)-1 {
				a.cursor++
			}
		case " ":
			a.targets[a.cursor].Selected = !a.targets[a.cursor].Selected
		case "a", "A":
			for i := range a.targets {
				a.targets[i].Selected = true
			}
		case "n", "N":
			for i := range a.targets {
				a.targets[i].Selected = false
			}
		case "s", "S":
			a.scanTargets()
			return
		}
	}
}

func (a *app) scanTargets() {
	a.term.PrintScanning("Calculating sizes...")

	hasSelection := false
	for _, t := range a.targets {
		if t.Selected {
			hasSelection = true
			break
		}
	}
	if !hasSelection {
		return
	}

	for i := range a.targets {
		if !a.targets[i].Selected {
			continue
		}
		a.term.PrintScanning(fmt.Sprintf("Scanning: %s", a.targets[i].Name))
		size := a.scanner.CalculateSizeForTarget(&a.targets[i])
		a.targets[i].Size = size
	}

	a.cursor = 0
	for {
		key := a.term.PrintResults(a.targets, a.cursor)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			if a.cursor < len(a.targets)-1 {
				a.cursor++
			}
		case " ":
			a.targets[a.cursor].Selected = !a.targets[a.cursor].Selected
		case "r", "R":
			a.scanTargets()
			return
		case "c", "C":
			if models.HasSelection(a.targets) {
				a.confirmAndClean()
				return
			}
		}
	}
}

func (a *app) confirmAndClean() {
	key := a.term.PrintConfirm(a.targets)
	switch key {
	case "y", "Y":
		a.cleanTargets()
	}
}

func (a *app) cleanTargets() {
	a.term.PrintCleaning("Starting cleanup...")

	results, totalSaved := a.cleaner.CleanTargets(a.targets, func(status string) {
		a.term.PrintCleaning(status)
	})

	// Collect errors
	var errorDetails []string
	for _, r := range results {
		if r.Error != nil {
			errorDetails = append(errorDetails, fmt.Sprintf("%s: %v", r.Target, r.Error))
		}
	}

	lastError := ""
	if len(errorDetails) > 0 {
		lastError = strings.Join(errorDetails, "\n")
	}

	for {
		key := a.term.PrintDone(totalSaved, lastError)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		}
	}
}

func (a *app) runBigFiles() {
	// Config
	key := a.term.PrintBigFilesConfig()
	var minSize int64
	switch key {
	case "1":
		minSize = 100 * 1024 * 1024
	case "2":
		minSize = 500 * 1024 * 1024
	case "3":
		minSize = 1024 * 1024 * 1024
	case "4":
		minSize = 5 * 1024 * 1024 * 1024
	case "b", "B", "q", "Q":
		return
	default:
		return
	}

	// Scan
	a.term.PrintScanning("Scanning for large files...")
	a.bigFiles = a.scanner.ScanBigFiles(minSize, func(status string) {
		a.term.PrintScanning(status)
	})

	// Sort by size
	sort.Slice(a.bigFiles, func(i, j int) bool {
		return a.bigFiles[i].Size > a.bigFiles[j].Size
	})

	a.selections = make(map[int]bool)
	a.cursor = 0

	// Results
	for {
		key := a.term.PrintBigFilesResults(a.bigFiles, a.selections, a.cursor, minSize)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			if a.cursor < len(a.bigFiles)-1 {
				a.cursor++
			}
		case " ":
			if len(a.bigFiles) > 0 {
				a.selections[a.cursor] = !a.selections[a.cursor]
			}
		case "a", "A":
			for i := range a.bigFiles {
				a.selections[i] = true
			}
		case "d", "D":
			if models.HasBigFilesSelection(a.selections) {
				a.deleteBigFiles()
				return
			}
		}
	}
}

func (a *app) deleteBigFiles() {
	a.term.PrintCleaning("Deleting files...")
	totalDeleted := a.cleaner.DeleteBigFiles(a.bigFiles, a.selections, func(status string) {
		a.term.PrintCleaning(status)
	})

	for {
		key := a.term.PrintDone(totalDeleted, "")
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		}
	}
}

func (a *app) runDuplicates() {
	key := a.term.PrintDuplicatesConfig()
	switch key {
	case "s", "S":
		a.scanDuplicates()
	case "b", "B":
		return
	case "q", "Q":
		os.Exit(0)
	}
}

func (a *app) scanDuplicates() {
	a.term.PrintScanning("Scanning for duplicates...")
	groups, _ := a.scanner.ScanDuplicates(func(status string) {
		a.term.PrintScanning(status)
	})
	a.duplicateGroups = groups
	a.selections = make(map[int]bool)
	a.cursor = 0

	// Show results and allow selection
	for {
		key := a.term.PrintDuplicatesResults(a.duplicateGroups, a.selections, a.cursor)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			if a.cursor < len(a.duplicateGroups)-1 {
				a.cursor++
			}
		case " ":
			if len(a.duplicateGroups) > 0 {
				a.selections[a.cursor] = !a.selections[a.cursor]
			}
		case "d", "D":
			if models.HasDuplicateSelection(a.selections) {
				a.deleteDuplicates()
				return
			}
		}
	}
}

func (a *app) deleteDuplicates() {
	a.term.PrintCleaning("Deleting duplicate files...")
	totalDeleted := a.cleaner.DeleteDuplicates(a.duplicateGroups, a.selections, func(status string) {
		a.term.PrintCleaning(status)
	})

	for {
		key := a.term.PrintDone(totalDeleted, "")
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		}
	}
}

func (a *app) runOldFiles() {
	key := a.term.PrintOldFilesConfig()
	var days int
	switch key {
	case "1":
		days = 30
	case "2":
		days = 90
	case "3":
		days = 180
	case "4":
		days = 365
	case "b", "B":
		return
	case "q", "Q":
		os.Exit(0)
	default:
		return
	}

	a.term.PrintScanning(fmt.Sprintf("Scanning for files > %d days old...", days))
	a.oldFiles = a.scanner.ScanOldFiles(days, func(status string) {
		a.term.PrintScanning(status)
	})
	a.selections = make(map[int]bool)
	a.cursor = 0

	// Show results and allow selection
	for {
		key := a.term.PrintOldFilesResults(a.oldFiles, a.selections, a.cursor, days)
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			if a.cursor < len(a.oldFiles)-1 {
				a.cursor++
			}
		case " ":
			if len(a.oldFiles) > 0 {
				a.selections[a.cursor] = !a.selections[a.cursor]
			}
		case "a", "A":
			for i := range a.oldFiles {
				a.selections[i] = true
			}
		case "d", "D":
			if models.HasOldFilesSelection(a.selections) {
				a.deleteOldFiles()
				return
			}
		}
	}
}

func (a *app) deleteOldFiles() {
	a.term.PrintCleaning("Deleting old files...")
	totalDeleted := a.cleaner.DeleteOldFiles(a.oldFiles, a.selections, func(status string) {
		a.term.PrintCleaning(status)
	})

	for {
		key := a.term.PrintDone(totalDeleted, "")
		switch key {
		case "q", "Q":
			os.Exit(0)
		case "b", "B":
			return
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	app := newApp()
	app.run()
}
