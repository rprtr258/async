package main

import (
	"encoding/base64"
	"encoding/gob"
	"os"
	"time"

	. "sems/http"
	. "sems/std"
)

type Session struct {
	ID     int
	Expiry time.Time
}

const OneWeek = time.Hour * 24 * 7

const SessionsFile = "sessions.gob"

var Sessions = make(map[string]*Session)

func GetSessionFromToken(token string) (*Session, error) {
	session, ok := Sessions[token]
	if !ok {
		return nil, NewError("session for this token does not exist")
	}

	now := time.Now()
	if session.Expiry.Before(now) {
		delete(Sessions, token)
		return nil, NewError("session for this token has expired")
	}

	session.Expiry = now.Add(OneWeek)
	return session, nil
}

func GetSessionFromRequest(r *HTTPRequest) (*Session, error) {
	return GetSessionFromToken(r.Cookie("Token"))
}

func GenerateSessionToken() (string, error) {
	const n = 64
	buffer := make([]byte, n)

	/* NOTE: see encoding/base64/base64.go:294. */
	token := make([]byte, (n+2)/3*4)

	if _, err := Getrandom(buffer, 0); err != nil {
		return "", err
	}

	base64.StdEncoding.Encode(token, buffer)

	return string(token), nil
}

func StoreSessionsToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(Sessions)
}

func RestoreSessionsFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return gob.NewDecoder(f).Decode(&Sessions)
}
