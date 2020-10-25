package stat

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
	"time"
)

/*
&{
    Dev:16777220
    Mode:33188
    Nlink:1
    Ino:58059237
    Uid:501
    Gid:20
    Rdev:0
    Pad_cgo_0: [0 0 0 0]
    Atimespec:{Sec:1603563201 Nsec:115934498}
    Mtimespec:{Sec:1603387176 Nsec:439532114}
    Ctimespec:{Sec:1603387176 Nsec:439532114}
    Birthtimespec:{Sec:1603387134 Nsec:447378890}
    Size:11358
    Blocks:24
    Blksize:4096
    Flags:0
    Gen:0
    Lspare:0
    Qspare:[0 0]
}
*/
type PlatformStat struct {
	inner *syscall.Stat_t
}

func NewPlatformStat(f os.FileInfo) PlatformStat {
	return PlatformStat{f.Sys().(*syscall.Stat_t)}
}

func (s PlatformStat) Links() uint16 {
	return s.inner.Nlink
}

func (s PlatformStat) INode() uint64 {
	return s.inner.Ino
}

func (s PlatformStat) Uid() uint32 {
	return s.inner.Uid
}

func (s PlatformStat) Username() string {
	username := fmt.Sprint(s.inner.Uid)
	owner, err := user.LookupId(username)

	if err == nil {
		username = owner.Username
	}

	return username
}

func (s PlatformStat) Gid() uint32 {
	return s.inner.Gid
}

func (s PlatformStat) Group() string {
	groupname := fmt.Sprint(s.inner.Gid)
	group, err := user.LookupGroupId(groupname)

	if err == nil {
		groupname = group.Name
	}

	return groupname
}

func (s PlatformStat) ATime() time.Time {
	return time.Unix(s.inner.Atimespec.Sec, s.inner.Atimespec.Nsec)
}

func (s PlatformStat) ModTime() time.Time {
	return time.Unix(s.inner.Mtimespec.Sec, s.inner.Mtimespec.Nsec)
}

func (s PlatformStat) CTime() time.Time {
	return time.Unix(s.inner.Ctimespec.Sec, s.inner.Ctimespec.Nsec)
}

func (s PlatformStat) Size() int64 {
	return s.inner.Size
}

func (s PlatformStat) Blocks() int64 {
	return s.inner.Blocks
}

func (s PlatformStat) BlockSize() int32 {
	return s.inner.Blksize
}
