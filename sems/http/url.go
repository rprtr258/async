package http

type URL struct {
	Path  string
	Query string
}

func ParseHexDigit(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'A' && c <= 'F':
		return 10 + c - 'A', true
	default:
		return 0, false
	}
}

func URLDecode(decoded []byte, encoded string) (int, bool) {
	var n int
	for i := 0; i < len(encoded); i++ {
		if encoded[i] == '%' {
			hi, ok := ParseHexDigit(encoded[i+1])
			if !ok {
				return 0, false
			}

			lo, ok := ParseHexDigit(encoded[i+2])
			if !ok {
				return 0, false
			}

			decoded[n] = byte(hi<<4 | lo)
			i += 2
		} else if encoded[i] == '+' {
			decoded[n] = ' '
		} else {
			decoded[n] = encoded[i]
		}
		n++
	}
	return n, true
}
