package imhttp

import "unicode"

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func parseU64(sv []byte) uint64 {
	var result uint64 = 0
	for i := 0; i < len(sv) && unicode.IsDigit(rune(sv[i])); i++ {
		result = result*10 + uint64(sv[i]-'0')
	}
	return result
}

func parseU64Hex(sv []byte) uint64 {
	var result uint64 = 0
	for _, x := range sv {
		var digit byte
		switch {
		case '0' <= x && x <= '9':
			digit = x - '0'
		case 'a' <= x && x <= 'z':
			digit = x - 'a' + 10
		case 'A' <= x && x <= 'Z':
			digit = x - 'A' + 10
		default:
			assert(false, "")
		}
		result = result*16 + uint64(digit)
	}
	return result
}
