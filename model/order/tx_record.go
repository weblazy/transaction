package order

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var TxRecordHandler = &TxRecord{}

// 事务记录表
type TxRecord struct {
	Id      int64 `json:"id" gorm:"primary_key;type:int AUTO_INCREMENT"`
	TxId    int64 `json:"tx_id" gorm:"column:tx_id;type:int NOT NULL;default:0;comment:'tx_id';index"`
	MysqlId int64 `json:"mysql_id" gorm:"column:mysql_id;type:int NOT NULL;default:0;comment:'mysql_id';index"`
	Status  int64 `json:"status" gorm:"column:status;type:int NOT NULL;default:0;comment:'status';index"`

	//正常配置
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*TxRecord) TableName() string {
	return "tx_record"
}

func (*TxRecord) Insert(db *gormx.DB, data *TxRecord) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*TxRecord) GetOne(where string, args ...interface{}) (*TxRecord, error) {
	var obj TxRecord
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*TxRecord) GetListWithLimit(limit int, where string, args ...interface{}) ([]*TxRecord, error) {
	var list []*TxRecord
	db := Orm()
	return list, db.Where(where, args...).Limit(limit).Find(&list).Error
}

func (*TxRecord) GetList(where string, args ...interface{}) ([]*TxRecord, error) {
	var list []*TxRecord
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*TxRecord) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&TxRecord{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*TxRecord) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&TxRecord{}).Error
}

func (*TxRecord) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&TxRecord{}).Where(where, args...).Update(data).Error
}
