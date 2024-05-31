package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Users struct {
	UserName string
	Password string
	Id       int
}

type Admins struct {
	UserName string
	Password string
	Id       int
}

type News struct {
	Time    time.Time
	Title   string
	Content string
	Author  string
	Id      int
	IsShow  bool
}

type Comment struct {
	Timestamp time.Time
	Content   string
	ID        int
	UserID    int
	NewsID    int
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
