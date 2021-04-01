package tx

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var TxHandler = &Tx{}

// 事务记录表
type Tx struct {
	Id     int64 `json:"id" gorm:"primary_key;type:int AUTO_INCREMENT"`
	Uid    int64 `json:"uid" gorm:"column:uid;type:int NOT NULL;default:0;comment:'用户id';index"`
	Status int64 `json:"status" gorm:"column:status;type:int NOT NULL;default:0;comment:'0未支付,1已支付';index"`
	//正常配置
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Tx) TableName() string {
	return "tx"
}

func (*Tx) Insert(db *gormx.DB, data *Tx) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Tx) GetOne(where string, args ...interface{}) (*Tx, error) {
	var obj Tx
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Tx) GetListWithLimit(limit int, where string, args ...interface{}) ([]*Tx, error) {
	var list []*Tx
	db := Orm()
	return list, db.Where(where, args...).Limit(limit).Find(&list).Error
}

func (*Tx) GetList(where string, args ...interface{}) ([]*Tx, error) {
	var list []*Tx
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Tx) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Tx{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Tx) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Tx{}).Error
}

func (*Tx) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) (int64, error) {
	if db == nil {
		db = Orm()
	}
	db = db.Model(&Tx{}).Where(where, args...).Update(data)
	return db.RowsAffected, db.Error
}
