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
	Admins   Admins     `gorm:"foreignKey:UID; references:ID;constraint:OnDelete:CASCADE;"`
	News     []News     `gorm:"foreignKey:UID;constraint:OnDelete:CASCADE;"`
	Comments []Comments `gorm:"foreignKey:UID;constraint:OnDelete:CASCADE;"`
	RateNews []RateNews `gorm:"foreignKey:UID;constraint:OnDelete:CASCADE;"`
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
	Comments  []Comments `gorm:"foreignKey:NID;constraint:OnDelete:CASCADE;"`
	RateNews  []RateNews `gorm:"foreignKey:NID;constraint:OnDelete:CASCADE;"`
}

type Comments struct {
	ID  int
	UID int
	NID int
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
	Rate      float32
	Timestamp time.Time
}

type RateNews struct {
	ID   int
	UID  int
	NID  int
	Rate int
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// 获取密码哈希
func getPasswordHash(password string) string {
	salt := "this is my salt"
	hashInstance := sha256.New()
	hashInstance.Write([]byte(password))
	tmp := hashInstance.Sum(nil)
	hashInstance.Reset()
	hashInstance.Write([]byte(hex.EncodeToString(tmp) + salt))
	return hex.EncodeToString(hashInstance.Sum(nil))
}

// 获取md5
func getMd5(content string) string {
	hashInstance := md5.New()
	hashInstance.Write([]byte(content))
	return hex.EncodeToString(hashInstance.Sum(nil))
}

// 获取文章摘要
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
