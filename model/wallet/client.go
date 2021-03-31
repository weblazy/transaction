package wallet

import (
	"github.com/jinzhu/gorm"
	g "github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

func Orm() *gorm.DB {
	db := g.GetORM("DbWallet")
	if utils.GetRunTime() != "onl" {
		db = db.Debug()
	}
	return db
}

func CreateTable() {
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='钱包表' AUTO_INCREMENT=1;").AutoMigrate(&Wallet{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='事务表' AUTO_INCREMENT=1;").AutoMigrate(&TxRecord{})
}
