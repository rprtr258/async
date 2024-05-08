package main

import (
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	. "sems/http"
	. "sems/std"
)

func findCharReverse(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}

	return -1
}

func DisplayShortenedString(w *HTTPResponse, s string, maxVisibleLen int) {
	if utf8.RuneCountInString(s) < maxVisibleLen {
		w.WriteHTMLString(s)
	} else {
		space := findCharReverse(s[:maxVisibleLen], ' ')
		if space == -1 {
			w.WriteHTMLString(s[:maxVisibleLen])
		} else {
			w.WriteHTMLString(s[:space])
		}
		w.AppendString(`...`)
	}
}

func GetIDFromURL(u URL, prefix string) (int, error) {
	path := u.Path

	if !strings.HasPrefix(path, prefix) {
		return 0, NotFound("requested page does not exist")
	}

	id, err := strconv.Atoi(path[len(prefix):])
	if err != nil {
		return 0, BadRequest("invalid ID for %q", prefix)
	}

	return id, nil
}

func GetIndicies(indicies string) (pindex int, spindex string, sindex int, ssindex string, err error) {
	if len(indicies) == 0 {
		return
	}

	spindex = indicies
	if i := strings.IndexByte(indicies, '.'); i != -1 {
		ssindex = indicies[i+1:]
		sindex, err = strconv.Atoi(ssindex)
		if err != nil {
			return
		}
		spindex = indicies[:i]
	}
	pindex, err = strconv.Atoi(spindex)
	return
}

func GetValidIndex[T any](si string, ts []T) (int, error) {
	i, err := strconv.Atoi(si)
	if err != nil {
		return 0, err
	}

	if i < 0 || i >= len(ts) {
		return 0, NewError("slice index out of range")
	}

	return i, nil
}

func MoveDown[T any](vs []T, i int) {
	if i >= 0 && i < len(vs)-1 {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveUp[T any](vs []T, i int) {
	if i > 0 && i <= len(vs)-1 {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func RemoveAtIndex[T any](ts []T, i int) []T {
	if i < 0 || i >= len(ts) {
		return ts
	}

	if i < len(ts)-1 {
		copy(ts[i:], ts[i+1:])
	}
	return ts[:len(ts)-1]
}

func StringLengthInRange(s string, min, max int) bool {
	runes := utf8.RuneCountInString(s)
	return runes >= min && runes <= max
}

func SetInt(vs url.Values, key string, value int) {
	vs.Set(key, strconv.Itoa(value))
}
