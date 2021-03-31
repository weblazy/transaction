package wallet

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var WalletHandler = &Wallet{}

// Wallet 钱包余额表
type Wallet struct {
	Id    int64 `json:"id" gorm:"primary_key;type:int AUTO_INCREMENT"`
	Money int64 `json:"money" gorm:"column:money;type:int NOT NULL;default:0;comment:'钱包余额';index"`
	//正常配置
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Wallet) TableName() string {
	return "wallet"
}

func (*Wallet) Insert(db *gormx.DB, data *Wallet) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Wallet) GetOne(where string, args ...interface{}) (*Wallet, error) {
	var obj Wallet
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Wallet) GetListWithLimit(limit int, where string, args ...interface{}) ([]*Wallet, error) {
	var list []*Wallet
	db := Orm()
	return list, db.Where(where, args...).Limit(limit).Find(&list).Error
}

func (*Wallet) GetList(where string, args ...interface{}) ([]*Wallet, error) {
	var list []*Wallet
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Wallet) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Wallet{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Wallet) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Wallet{}).Error
}

func (*Wallet) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&Wallet{}).Where(where, args...).Update(data).Error
}
