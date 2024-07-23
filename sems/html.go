package main

import (
	"time"

	. "github.com/rprtr258/async/sems/http"
)

func DisplayFormattedTime(w *HTTPResponse, t time.Time) {
	w.Write(t.AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
}

func DisplayConstraintInput(w *HTTPResponse, t string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <input type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintIndexedInput(w *HTTPResponse, t string, minLength, maxLength int, name string, index int, value string, required bool) {
	w.AppendString(` <input type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintTextarea(w *HTTPResponse, cols, rows string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <textarea cols="`)
	w.AppendString(cols)
	w.AppendString(`" rows="`)
	w.AppendString(rows)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
	w.WriteHTMLString(value)
	w.AppendString(`</textarea>`)
}

func DisplayIndexedCommand(w *HTTPResponse, index int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.AppendString(command)
	w.AppendString(`" formnovalidate>`)
}

func DisplayDoublyIndexedCommand(w *HTTPResponse, pindex, sindex int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(pindex)
	w.AppendString(`.`)
	w.WriteInt(sindex)
	w.AppendString(`" value="`)
	w.AppendString(command)
	w.AppendString(`" formnovalidate>`)
}

func DisplayErrorMessage(w *HTTPResponse, message string) {
	if message != "" {
		w.AppendString(`<div><p>Error: `)
		w.WriteHTMLString(message)
		w.AppendString(`.</p></div>`)
	}
}

func ErrorPageHandler(w *HTTPResponse, message string) {
	w.Bodies = w.Bodies[:0]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Error</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Master's degree</h1>`)
	w.AppendString(`<h2>Error</h2>`)

	DisplayErrorMessage(w, message)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
}
