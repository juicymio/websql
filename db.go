package main

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os/exec"
)

func connectDb() (db *gorm.DB) {
	dsn := "root:root@tcp(127.0.0.1:3306)/websql?charset=utf8mb4&parseTime=True&loc=Local"
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
	err = db.AutoMigrate(&Comments{})
	checkErr(err)
	return db
}

func testDb(db *gorm.DB) {
	rows, err := db.Raw("SHOW databases").Rows()
	checkErr(err)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
	}
}

// TODO: Add frontend interface
func backupDb(dbName, username, password, backupFile string) error {
	cmd := fmt.Sprintf("mysqldump -u%s -p%s %s > %s", username, password, dbName, backupFile)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}
	return nil
}

func restoreDb(dbName, username, password, backupFile string) error {
	cmd := fmt.Sprintf("mysql -u%s -p%s %s < %s", username, password, dbName, backupFile)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}
	return nil
}

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

func getAllNews(db *gorm.DB) []News {
	var news []News
	db.Find(&news)
	return news
}

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

func getComments(db *gorm.DB, NID string) ([]Comments, error) {
	var comments []Comments
	res := db.Where("n_id = ?", NID).Find(&comments)
	return comments, res.Error
}

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

func addLike(db *gorm.DB, like LikeComment) error {
	res := db.Create(&like)
	return res.Error
}

func getRate(db *gorm.DB, NID int, UID int) (int, error) {
	var rate RateNews
	res := db.Where("n_id = ? AND uid = ?", NID, UID).First(&rate)
	if res.Error != nil {
		return -1, res.Error
	} else {
		return rate.Rate, nil
	}
}

func getAverageRate(db *gorm.DB, NID int) (float32, error) {
	var rates []RateNews
	db.Where("n_id = ?", NID).Find(&rates)
	if len(rates) == 0 {
		return 2.5, nil
	} else {
		var total float32
		for _, rate := range rates {
			total += float32(rate.Rate)
		}
		return total / float32(len(rates)), nil
	}
}

func getOrderNews(db *gorm.DB) ([]News, error) {
	var news []News
	db.Raw("SELECT n.*, AVG(r.rate) AS avg_rate FROM news n LEFT JOIN rate_news r ON n.id = r.n_id GROUP BY n.id ORDER BY avg_rate DESC").Scan(&news)
	return news, nil
}
