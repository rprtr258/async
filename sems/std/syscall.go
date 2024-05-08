package std

import (
	"syscall"
	"unsafe"
)

const (
	/* See <sys/syscall.h>. */
	// I/O
	SYS_read   = syscall.SYS_READ
	SYS_write  = syscall.SYS_WRITE
	SYS_writev = syscall.SYS_WRITEV
	// Process
	SYS_kill = syscall.SYS_KILL
	SYS_exit = syscall.SYS_EXIT
	// Filesystem
	SYS_open      = syscall.SYS_OPEN
	SYS_close     = syscall.SYS_CLOSE
	SYS_access    = syscall.SYS_ACCESS
	SYS_mkdir     = syscall.SYS_MKDIR
	SYS_rmdir     = syscall.SYS_RMDIR
	SYS_stat      = syscall.SYS_STAT
	SYS_lseek     = syscall.SYS_LSEEK
	SYS_ftruncate = syscall.SYS_FTRUNCATE
	SYS_fstat     = syscall.SYS_FSTAT
	// Socket
	SYS_accept     = syscall.SYS_ACCEPT
	SYS_fcntl      = syscall.SYS_FCNTL
	SYS_socket     = syscall.SYS_SOCKET
	SYS_bind       = syscall.SYS_BIND
	SYS_setsockopt = syscall.SYS_SETSOCKOPT
	SYS_listen     = syscall.SYS_LISTEN
	// Other
	SYS_clock_gettime = syscall.SYS_CLOCK_GETTIME
	SYS_getrandom     = 318
	SYS_mmap          = syscall.SYS_MMAP
	SYS_nanosleep     = syscall.SYS_NANOSLEEP
	SYS_shm_open2     = syscall.SYS_SHMCTL
	SYS_unlink        = syscall.SYS_UNLINK
	SYS_unmount       = syscall.SYS_UMOUNT2
)

type FD = int

func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, errno syscall.Errno) {
	return syscall.RawSyscall(trap, a1, a2, a3)
}

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, errno syscall.Errno) {
	return RawSyscall(trap, a1, a2, a3)
}

func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, errno syscall.Errno) {
	return syscall.RawSyscall6(trap, a1, a2, a3, a4, a5, a6)
}

func Accept(s FD, addr *SockAddr, addrlen *uint32) (FD, error) {
	r1, _, errno := Syscall(SYS_accept, uintptr(s), uintptr(unsafe.Pointer(addr)), uintptr(unsafe.Pointer(addrlen)))
	return FD(r1), NewSyscallError("accept failed with code", errno)
}

func Access(path string, mode int32) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_access, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), uintptr(mode), 0)
	return NewSyscallError("access failed with code", syscall.Errno(errno))
}

func Bind(s FD, addr *SockAddr, addrlen uint32) error {
	_, _, errno := RawSyscall(SYS_bind, uintptr(s), uintptr(unsafe.Pointer(addr)), uintptr(addrlen))
	return NewSyscallError("bind failed with code", syscall.Errno(errno))
}

func ClockGettime(clockID int32, tp *Timespec) error {
	_, _, errno := RawSyscall(SYS_clock_gettime, uintptr(clockID), uintptr(unsafe.Pointer(tp)), 0)
	return NewSyscallError("clock_gettime failed with code", errno)
}

func Close(fd FD) error {
	_, _, errno := Syscall(SYS_close, uintptr(fd), 0, 0)
	return NewSyscallError("close failed with code", errno)
}

func Exit(status int32) {
	RawSyscall(SYS_exit, uintptr(status), 0, 0)
}

func Fcntl(fd FD, cmd, arg int32) error {
	_, _, errno := Syscall(SYS_fcntl, uintptr(fd), uintptr(cmd), uintptr(arg))
	return NewSyscallError("fcntl failed with code", errno)
}

func Fstat(fd FD, sb *Stat_t) error {
	_, _, errno := RawSyscall(SYS_fstat, uintptr(fd), uintptr(unsafe.Pointer(sb)), 0)
	return NewSyscallError("fstat failed with code", errno)
}

func Ftruncate(fd FD, length int64) error {
	_, _, errno := RawSyscall(SYS_ftruncate, uintptr(fd), uintptr(length), 0)
	return NewSyscallError("ftruncate failed with code", errno)
}

func Getrandom(buf []byte, flags uint32) (int64, error) {
	r1, _, errno := Syscall(SYS_getrandom, uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)), uintptr(flags))
	return int64(r1), NewSyscallError("getrandom failed with code", errno)
}

func Listen(s FD, backlog int32) error {
	_, _, errno := RawSyscall(SYS_listen, uintptr(s), uintptr(backlog), 0)
	return NewSyscallError("listen failed with code", errno)
}

func Lseek(fd FD, offset int64, whence int32) (int64, error) {
	r1, _, errno := RawSyscall(SYS_lseek, uintptr(fd), uintptr(offset), uintptr(whence))
	return int64(r1), NewSyscallError("lseek failed with code", errno)
}

func Mkdir(path string, mode int16) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_mkdir, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), uintptr(mode), 0)
	return NewSyscallError("mkdir failed with code", errno)
}

func Mmap(addr unsafe.Pointer, len uint64, prot, flags int32, fd FD, offset int64) (unsafe.Pointer, error) {
	r1, _, errno := Syscall6(SYS_mmap, uintptr(addr), uintptr(len), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	return unsafe.Pointer(r1), NewSyscallError("mmap failed with code", errno)
}

func Nanosleep(rqtp, rmtp *Timespec) error {
	_, _, errno := Syscall(SYS_nanosleep, uintptr(unsafe.Pointer(rqtp)), uintptr(unsafe.Pointer(rmtp)), 0)
	return NewSyscallError("nanosleep failed with code", errno)
}

func Open(path string, flags int32, mode uint16) (int32, error) {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	r1, _, errno := Syscall(SYS_open, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), uintptr(flags), uintptr(mode))
	return int32(r1), NewSyscallError("open failed with code", errno)
}

func Read(fd FD, buf []byte) (int64, error) {
	r1, _, errno := Syscall(SYS_read, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))
	return int64(r1), NewSyscallError("read failed with code", errno)
}

func Rmdir(path string) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_rmdir, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), 0, 0)
	return NewSyscallError("rmdir failed with code", errno)
}

func Setsockopt(s FD, level, optname int32, optval unsafe.Pointer, optlen uint32) error {
	_, _, errno := Syscall6(SYS_setsockopt, uintptr(s), uintptr(level), uintptr(optname), uintptr(optval), uintptr(optlen), 0)
	return NewSyscallError("setsockopt failed with code", errno)
}

func ShmOpen2(path string, flags int32, mode uint16, shmflags int32, name string) (int, error) {
	r1, _, errno := Syscall6(SYS_shm_open2, uintptr(unsafe.Pointer(unsafe.StringData(path))), uintptr(flags), uintptr(mode), uintptr(shmflags), uintptr(unsafe.Pointer(unsafe.StringData(name))), 0)
	return int(r1), NewSyscallError("shm_open2 failed with code", errno)
}

func Socket(domain, typ, protocol int32) (FD, error) {
	r1, _, errno := RawSyscall(SYS_socket, uintptr(domain), uintptr(typ), uintptr(protocol))
	return FD(r1), NewSyscallError("socket failed with code", errno)
}

func Stat(path string, sb *Stat_t) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_stat, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), uintptr(unsafe.Pointer(sb)), 0)
	return NewSyscallError("stat failed with code", errno)
}

func Unlink(path string) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_unlink, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), 0, 0)
	return NewSyscallError("unlink failed with code", errno)
}

func Unmount(path string, flags int32) error {
	buffer := make([]byte, PATH_MAX)
	n := copy(buffer, path)

	_, _, errno := RawSyscall(SYS_unmount, uintptr(unsafe.Pointer(unsafe.SliceData(buffer[:n+1]))), uintptr(flags), 0)
	return NewSyscallError("unmount failed with code", errno)
}

func Write(fd FD, buf []byte) (int64, error) {
	r1, _, errno := Syscall(SYS_write, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))
	return int64(r1), NewSyscallError("write failed with code", errno)
}

func Writev(fd FD, iov []Iovec) (int64, error) {
	r1, _, errno := Syscall(SYS_writev, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(iov))), uintptr(len(iov)))
	return int64(r1), NewSyscallError("writev failed with code", errno)
}
