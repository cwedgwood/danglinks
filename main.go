//

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

const (
	EX_USAGE = 64
)

var (
	optDebug    bool
	optZeroTerm bool
	optCrossDev bool
	optRepeat   bool
	optSilence  bool
	optVerbose  bool

	mpdev      uint64
	seenBefore = make(map[uint64]bool)
)

func main() {

	flag.BoolVar(&optDebug, "d", false, "Debug mode (not that useful)")
	flag.BoolVar(&optZeroTerm, "0", false, "Zero-terminate output")
	flag.BoolVar(&optZeroTerm, "z", false, "Zero-terminate output")
	flag.BoolVar(&optCrossDev, "x", false, "Cross devices (mountpoints)")
	flag.BoolVar(&optRepeat, "r", false, "Repeat inodes we've encountered before")
	flag.BoolVar(&optSilence, "s", false, "Silence traversal errors")
	flag.BoolVar(&optVerbose, "v", false, "Verbose output (incompatible with zero-terminated)")

	flag.Parse()

	if optVerbose && optZeroTerm {
		fmt.Fprintf(os.Stderr, "Cannot be zero-terminated and verbose at the same time.\n")
		os.Exit(EX_USAGE)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		Walk(path)
	}

}

func errOut(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func debugOut(format string, args ...any) {
	if !optDebug {
		return
	}
	errOut(format, args...)
}

func Walk(searchPath string) {

	debugOut("Searching: %s", searchPath)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			if optSilence {
				return nil
			}
			errOut("Traversal error: %s\n", err)
			return nil
		}

		st := info.Sys().(*syscall.Stat_t)
		if !optCrossDev && info.IsDir() {
			if mpdev == 0 {
				mpdev = st.Dev
			}
			if st.Dev != mpdev {
				return filepath.SkipDir
			}

		}

		// anything we've seen before, we'll skip; this includes directories and symlinks
		if !optRepeat {
			ino := st.Ino
			if seenBefore[ino] {
				if info.IsDir() {
					debugOut("Skipping directory: %s\n", path)
					return filepath.SkipDir
				}
				debugOut("Skipping path: %s\n", path)
				return nil
			}
			seenBefore[ino] = true
		}

		// only interested in symlinks
		if info.Mode()&fs.ModeSymlink == 0 {
			return nil
		}

		// danglingg?
		_, err = os.Stat(path)
		if err == nil {
			// no, move on
			return nil
		}

		// 0-terminated output
		if optZeroTerm {
			fmt.Printf("%s\x00", path)
			return nil
		}

		// normal output
		if !optVerbose {
			fmt.Printf("%s\n", path)
			return nil
		}

		// verbose output
		target, err := os.Readlink(path)
		if err != nil {
			errOut("Readlink Error: %v\n", err)
			return nil
		}

		fmt.Printf("%s → %s\n", path, target)
		return nil
	})

	if err != nil {
		errOut("Walk error: %v\n", err)
	}
}
