package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Users struct {
	ID       int
	UserName string
	Password string
}

type Admins struct {
	ID  int
	UID int
}

type News struct {
	ID        int
	UID       int
	Title     string
	Content   string
	IsShow    bool
	Timestamp time.Time
}

type Comments struct {
	ID        int
	UID       int
	NID       int
	Content   string
	Timestamp time.Time
}

type Render struct {
	Author    string
	Content   string
	Timestamp time.Time
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getPasswordHash(password string) string {
	salt := "this is my salt"
	hashInstance := sha256.New()
	hashInstance.Write([]byte(password))
	bytes := hashInstance.Sum(nil)
	hashInstance.Reset()
	hashInstance.Write([]byte(hex.EncodeToString(bytes) + salt))
	return hex.EncodeToString(hashInstance.Sum(nil))
}
