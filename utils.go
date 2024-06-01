package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Users struct {
	ID       int
	UserName string
	Password string
}

type Admins struct {
	ID  int
	UID int // userID
}

type News struct {
	ID        int
	UID       int // userID
	Title     string
	Content   string
	IsShow    bool
	Timestamp time.Time
}

type Comments struct {
	ID  int
	UID int // userID
	NID int // newsID
	// FID       int // father
	Content   string
	Timestamp time.Time
}

type RenderComments struct {
	Author    string
	Content   string
	Timestamp time.Time
}

type RenderNews struct {
	ID        int
	Title     string
	Author    string
	Content   template.HTML
	Timestamp time.Time
}

type RateNews struct {
	ID   int
	UID  int
	NID  int
	Rate int
}

type LikeComment struct {
	ID    int
	UID   int
	CID   int
	Value bool
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

func getMd5(content string) string {
	hashInstance := md5.New()
	hashInstance.Write([]byte(content))
	return hex.EncodeToString(hashInstance.Sum(nil))
}

func truncateHTML(s string, maxLen int) string {
	z := html.NewTokenizer(strings.NewReader(s))

	var buf bytes.Buffer
	totalLen := 0

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		token := z.Token()
		if tt == html.TextToken {
			totalLen += len(token.String())
			if totalLen > maxLen {
				break
			}
		}
		buf.WriteString(token.String())
	}

	return buf.String()
}
