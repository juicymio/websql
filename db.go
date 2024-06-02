package main

import (
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectDb() (db *gorm.DB) {
	dsn := "root:889047ll@tcp(127.0.0.1:3306)/websql?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	checkErr(err)
	err = db.AutoMigrate(&Users{})
	checkErr(err)
	err = db.AutoMigrate(&Admins{})
	checkErr(err)
	err = db.AutoMigrate(&News{})
	checkErr(err)
	err = db.AutoMigrate(&Comments{})
	checkErr(err)
	err = db.AutoMigrate(&RateNews{})
	checkErr(err)
	return db
}

// func testDb(db *gorm.DB) {
// 	rows, err := db.Raw("SHOW databases").Rows()
// 	checkErr(err)
// 	for rows.Next() {
// 		var name string
// 		err = rows.Scan(&name)
// 	}
// }

// User
// 通过用户名查询密码并判断返回
func checkUser(db *gorm.DB, users Users) (int, error) {
	var getUser Users
	res := db.Select("password").Where("user_name = ?", users.UserName).First(&getUser)
	if res.Error != nil {
		return -1, res.Error
	}

	if getUser.Password != getPasswordHash(users.Password) {
		return -1, errors.New("wrong password")
	} else {
		uid, _ := userChange(db, -1, users.UserName)
		return uid, nil
	}
}

// 用户用户名和id互转
func userChange(db *gorm.DB, id int, username string) (int, string) {
	var user Users
	if username == "" {
		db.Select("user_name").Where("id = ?", id).First(&user)
		return -1, user.UserName
	} else {
		db.Select("id").Where("user_name = ?", username).First(&user)
		return user.ID, ""
	}
}

func register(db *gorm.DB, user Users) error {
	if !errors.Is(db.Where("user_name = ?", user.UserName).First(&Users{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("user already exists")
	}

	res := db.Create(&user)
	return res.Error
}

// 输入用户id更新信息
func updateUser(db *gorm.DB, user Users) error {
	if !errors.Is(db.Where("user_name = ?", user.UserName).First(&Users{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("user already exists")
	}

	res := db.Model(&user).Where("id = ?", user.ID).Updates(user)
	return res.Error
}

// Admin
// 返回是否为admin
func checkAdmin(db *gorm.DB, users Users) (int, error) {
	uid, err := checkUser(db, users)
	if err != nil {
		return -1, err
	}

	if errors.Is(db.Where("uid = ?", uid).First(&Admins{}).Error, gorm.ErrRecordNotFound) {
		return -1, errors.New("not Admin")
	} else {
		return uid, nil
	}
}

// News
func addNews(db *gorm.DB, news News) error {
	res := db.Create(&news)
	return res.Error
}

func updateNews(db *gorm.DB, news News) error {
	res := db.Model(&news).Where("id = ?", news.ID).Select("title", "content", "is_show").Updates(news)
	return res.Error
}

func deleteNews(db *gorm.DB, id int) error {
	res := db.Where("id = ?", id).Delete(&News{})
	return res.Error
}

// 依据时间获取所有文章排序
func getAllNews(db *gorm.DB) []News {
	var news []News
	db.Order("timestamp DESC").Find(&news)
	return news
}

// 依据id获取文章
func getNews(db *gorm.DB, id string) (News, error) {
	var news News
	res := db.Where("id = ?", id).First(&news)
	return news, res.Error
}

// Comments
func addComment(db *gorm.DB, comments Comments) error {
	res := db.Create(&comments)
	return res.Error
}

// 获取某一文章所有评论
func getComments(db *gorm.DB, NID string) ([]Comments, error) {
	var comments []Comments
	res := db.Where("n_id = ?", NID).Find(&comments)
	return comments, res.Error
}

// 更新对文章的评价，如果未平均就创建评价
func updateRate(db *gorm.DB, rate RateNews) error {
	if errors.Is(db.Where("n_id = ? AND uid = ?", rate.NID, rate.UID).First(&RateNews{}).Error, gorm.ErrRecordNotFound) {
		res := db.Create(&rate)
		return res.Error
	} else {
		fmt.Println(rate)
		res := db.Model(&rate).Where("n_id = ? AND uid = ?", rate.NID, rate.UID).Updates(rate)
		return res.Error
	}
}

// 搜索文章
func searchNews(db *gorm.DB, query string) ([]News, error) {
	var news []News

	likeQuery := "%" + query + "%"
	res := db.Joins("left join users on news.UID = users.ID").Where("news.Title LIKE ? OR news.Content LIKE ? OR users.user_name LIKE ?", likeQuery, likeQuery, likeQuery).Order("timestamp DESC").Find(&news)

	if res.Error != nil {
		return nil, res.Error
	}

	return news, nil
}

// 获取某一用户对某一文章的评价
func getRate(db *gorm.DB, NID int, UID int) (int, error) {
	var rate RateNews
	res := db.Where("n_id = ? AND uid = ?", NID, UID).First(&rate)
	if res.Error != nil {
		return -1, res.Error
	} else {
		return rate.Rate, nil
	}
}

// 获取某一文章的平均分
func getAverageRate(db *gorm.DB, NID int) (float32, error) {
	var rates []RateNews
	db.Where("n_id = ?", NID).Find(&rates)
	if len(rates) == 0 {
		return 0, nil
	} else {
		var total float32
		for _, rate := range rates {
			total += float32(rate.Rate)
		}
		return total / float32(len(rates)), nil
	}
}

// 依据平均分排序返回文章
func getOrderNews(db *gorm.DB) ([]News, error) {
	var news []News
	db.Raw("SELECT n.*, AVG(r.rate) AS avg_rate FROM news n LEFT JOIN rate_news r ON n.id = r.n_id GROUP BY n.id ORDER BY avg_rate DESC").Scan(&news)
	return news, nil
}
