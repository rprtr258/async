package std

import (
	"syscall"
)

type SockAddrIn struct {
	Len    uint8
	Family uint8
	Port   uint16
	Addr   uint32
	_      [8]byte
}

/* NOTE: actually SockAddr is the following structure:
 * struct sockaddr {
 *	unsigned char	sa_len;		// total length
 *	sa_family_t	sa_family;	// address family
 *	char		sa_data[14];	// actually longer; address value
 * };
 * But because I don't really care, and sizes are the same, I made them synonyms.
 */
type SockAddr = SockAddrIn

const (
	AF_INET = syscall.AF_INET

	SOCK_STREAM = syscall.SOCK_STREAM

	SOL_SOCKET = 0xFFFF

	SO_REUSEADDR    = syscall.SO_REUSEADDR
	SO_REUSEPORT    = 0x00000200
	SO_REUSEPORT_LB = 0x00010000
	SO_RCVTIMEO     = 0x00001006

	SHUT_RD = 0
	SHUT_WR = 1

	/* From <netinet/in.h>. */
	INADDR_ANY = 0
)

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func TCPListen(port uint16) (FD, error) {
	l, err := Socket(AF_INET, SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}

	if err := syscall.SetsockoptInt(l, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return -1, err
	}

	if err := syscall.Bind(l, &syscall.SockaddrInet4{Port: int(port), Addr: [4]byte{}}); err != nil {
		return -1, err
	}

	const backlog = 128
	if err := Listen(l, backlog); err != nil {
		return -1, err
	}

	return l, nil
}
