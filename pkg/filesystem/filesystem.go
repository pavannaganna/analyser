package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/pavannaganna/analyser/pkg/views"

	"golang.org/x/sys/unix"

	"code.cloudfoundry.org/bytefmt"
)

var db = make(map[string]string)

const (
	largeFilenameKey = "LARGE_FILE_NAME"
	largeSizeKey     = "LARGE_SIZE"
	totalSizeKey     = "TOTAL_SIZE"
	base             = 10
	bitSize          = 64
	floatPrec        = 6
)

// SpaceInputs defines the inputs to the space processing
type SpaceInputs struct {
	FilesystemPath string
	RootFilesystem string
	Filter         string
}

// DirStats describe stats about a directory
type DirStats struct {
	Path      string
	TotalSize int64
}

type disk struct {
	all     uint64
	free    uint64
	used    uint64
	usedPer float64
}

//Space analyses the space problems for a given filesystem path
func Space(inputs SpaceInputs) error {
	filesystemPath := inputs.FilesystemPath
	db["filter"] = inputs.Filter
	var err error
	if !verify(filesystemPath) {
		err = errors.New("Path is not abosule path")
		return err
	}
	start := time.Now()
	err = filepath.Walk(filesystemPath, process)
	root := getDiskUsage("/")
	if err != nil {
		return err
	}
	t := time.Now()
	elapsed := t.Sub(start)
	elapsedInMin := elapsed.Minutes()
	largeFileProcessed, _ := db[largeFilenameKey]
	largeSizeProcessed, _ := db[largeSizeKey]
	totalSizeProcessed, _ := db[totalSizeKey]
	largeSize, _ := strconv.ParseUint(largeSizeProcessed, base, bitSize)
	totalSize, _ := strconv.ParseUint(totalSizeProcessed, base, bitSize)
	if root.all > 0 {
		totalSize = root.all
	}
	largeFileSizePercentage := (float64(largeSize) / float64(totalSize)) * 100.00
	data := [][]string{
		[]string{"LARGE_FILE_NAME", largeFileProcessed, "NA"},
		[]string{"LARGE_FILE_SIZE", bytefmt.ByteSize(largeSize), strconv.FormatFloat(largeFileSizePercentage, 'f', floatPrec, bitSize)},
		[]string{"DISK_TOTAL_SIZE", bytefmt.ByteSize(root.all), "NA"},
		[]string{"DISK_USED_PERCENTAGE", "--", strconv.FormatFloat(root.usedPer, 'f', floatPrec, bitSize)},
		[]string{"PROCESSING_TIME", strconv.FormatFloat(elapsedInMin, 'f', floatPrec, bitSize) + " min(s)", "NA"},
	}
	views.Print(data)
	return err
}

// VolumeScanner scans a given volume and reports the usage per directory
// under the volume
func VolumeScanner(pathToScan string) error {
	result := make(map[string]DirStats)
	var largestDir string
	largestSize := int64(0)
	totalSize := int64(0)
	var cwd string
	err := filepath.Walk(pathToScan, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			cwd = path
			result[cwd] = DirStats{
				Path:      cwd,
				TotalSize: int64(0),
			}
			return err
		}
		dir := filepath.Dir(path)
		dirStat, _ := result[dir]
		dirStat.TotalSize += info.Size()
		totalSize += info.Size()
		if dirStat.TotalSize > largestSize {
			largestSize = dirStat.TotalSize
			largestDir = dir
		}
		return err

	})
	data := [][]string{
		[]string{"TOTAL_SIZE", bytefmt.ByteSize(uint64(totalSize)), ""},
		[]string{"LARGEST_DIR", largestDir, ""},
		[]string{"LARGEST_DIR_SIZE", bytefmt.ByteSize(uint64(largestSize)), ""},
	}

	views.Print(data)
	return err

}

func verify(pathToVerify string) bool {
	valid := false
	if _, err := os.Stat(pathToVerify); err == nil {
		valid = true
	}
	return valid
}

func process(filesystemPath string, info os.FileInfo, err error) error {
	isDir := info.IsDir()
	filter := db["filter"]
	if isDir {
		// we only find the file in a size having large size
		return err
	}

	if filter != "" {
		fileNameMatched := regexp.MustCompilePOSIX(filter).MatchString(info.Name())
		if !fileNameMatched {
			return err
		}
	}
	size := info.Size()
	var largeSize int64
	var totalSize int64
	largeSizeFromCache, ok := db[largeSizeKey]
	if !ok {
		largeSizeFromCache = "0"
	}
	totalSizeFromCache, ok := db[totalSizeKey]
	if !ok {
		totalSizeFromCache = "0"
	}
	totalSize, _ = strconv.ParseInt(totalSizeFromCache, 10, 64)
	largeSize, _ = strconv.ParseInt(largeSizeFromCache, 10, 64)
	if err != nil {
		return err
	}
	if largeSize < size {
		largeSize = size
		db[largeSizeKey] = strconv.FormatInt(largeSize, 10)
		db[largeFilenameKey] = filesystemPath
	}
	totalSize = size + totalSize
	db[totalSizeKey] = strconv.FormatInt(totalSize, 10)
	return err

}

func getDiskUsage(rootPath string) disk {
	fs := unix.Statfs_t{}
	err := unix.Statfs(rootPath, &fs)
	var root disk
	if err != nil {
		return root
	}
	// Available blocks * size per block = available space in bytes
	root.all = fs.Blocks * uint64(fs.Bsize)
	root.free = fs.Bfree * uint64(fs.Bsize)
	root.used = root.all - root.free
	root.usedPer = float64(root.used) / float64(root.all) * 100.00

	return root
}
