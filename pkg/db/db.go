package db

import (
	"fmt"
	"time"

	"reliproxy/pkg/repository"
	"reliproxy/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	maxRetries    = 5
	retryInterval = time.Second * 10
)

type MySQLConnectionEnv struct {
	Host     string
	Port     string
	User     string
	DBName   string
	Password string
}

func NewMySQLConnectionEnv() *MySQLConnectionEnv {
	return &MySQLConnectionEnv{
		Host:     utils.GetEnv("MYSQL_HOST", "localhost"),
		Port:     utils.GetEnv("MYSQL_PORT", "3306"),
		User:     utils.GetEnv("MYSQL_USER", "root"),
		DBName:   utils.GetEnv("MYSQL_DBNAME", "my_tutor"),
		Password: utils.GetEnv("MYSQL_PASSWORD", "password"),
	}
}

func (mc *MySQLConnectionEnv) connectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mc.User, mc.Password, mc.Host, mc.Port, mc.DBName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func (mc *MySQLConnectionEnv) ConnectDBWithRetry() (*gorm.DB, error) {
	var err error
	var db *gorm.DB

	for i := 0; i < maxRetries; i++ {
		db, err = mc.connectDB()
		if err == nil {
			return db, nil
		}
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("could not connect to MySQL after %d retries: %v", maxRetries, err)
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&repository.RequestStatus{})
	if err != nil {
		return fmt.Errorf("failed to migrate RequestStatus model: %v", err)
	}
	return nil
}
