package tx

import (
	"github.com/jinzhu/gorm"
	g "github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

func Orm() *gorm.DB {
	db := g.GetORM("DbTx")
	if utils.GetRunTime() != "onl" {
		db = db.Debug()
	}
	return db
}

func CreateTable() {
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='事务记录表' AUTO_INCREMENT=1;").AutoMigrate(&Tx{})
}
