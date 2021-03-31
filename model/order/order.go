package order

import (
	"time"

	gormx "github.com/jinzhu/gorm"
)

var OrderHandler = &Order{}

//Order 订单表
type Order struct {
	Id     int64 `json:"id" gorm:"primary_key;type:int AUTO_INCREMENT"`
	Status int64 `json:"status" gorm:"column:status;type:int NOT NULL;default:0;comment:'0未支付,1已支付';index"`
	//正常配置
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;NOT NULL;default:CURRENT_TIMESTAMP;type:TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;NOT NULL;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;type:TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at;type:DATETIME"`
}

func (*Order) TableName() string {
	return "order"
}

func (*Order) Insert(db *gormx.DB, data *Order) error {
	if db == nil {
		db = Orm()
	}
	return db.Create(data).Error
}

func (*Order) GetOne(where string, args ...interface{}) (*Order, error) {
	var obj Order
	return &obj, Orm().Where(where, args...).Take(&obj).Error
}

func (*Order) GetListWithLimit(limit int, where string, args ...interface{}) ([]*Order, error) {
	var list []*Order
	db := Orm()
	return list, db.Where(where, args...).Limit(limit).Find(&list).Error
}

func (*Order) GetList(where string, args ...interface{}) ([]*Order, error) {
	var list []*Order
	db := Orm()
	return list, db.Where(where, args...).Find(&list).Error
}

func (*Order) GetCount(where string, args ...interface{}) (int, error) {
	var number int
	err := Orm().Model(&Order{}).Where(where, args...).Count(&number).Error
	return number, err
}

func (*Order) Delete(db *gormx.DB, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Where(where, args...).Delete(&Order{}).Error
}

func (*Order) Update(db *gormx.DB, data map[string]interface{}, where string, args ...interface{}) error {
	if db == nil {
		db = Orm()
	}
	return db.Model(&Order{}).Where(where, args...).Update(data).Error
}
