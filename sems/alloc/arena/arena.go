package arena

type Arena struct {
	Buffer []byte
	Used   int
}

func (a *Arena) Alloc(n int) []byte {
	if a.Used+n >= cap(a.Buffer) {
		buffer := make([]byte, max(a.Used+n*2, cap(a.Buffer)*2))
		copy(buffer, a.Buffer)
		a.Buffer = buffer
	}
	a.Used += n
	return a.Buffer[a.Used-n : a.Used]
}

func (a *Arena) Reset() {
	a.Used = 0
}
