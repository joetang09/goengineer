package engineer

import (
	"os"
	"syscall"
	"time"
)

func FileCreateTime(fi os.FileInfo) time.Time {

	ad := fi.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(int64(ad.CreationTime.Nanoseconds()/1e9), 0)

}
