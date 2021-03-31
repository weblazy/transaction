package main

import (
	"fmt"
	"os"

	"gopkg.in/redis.v5"

	"github.com/sunmi-OS/gocore/gorm"
	"github.com/weblazy/gocore/utils"
	"github.com/weblazy/transaction/config"
	"github.com/weblazy/transaction/model/order"
	"github.com/weblazy/transaction/model/tx"
	"github.com/weblazy/transaction/model/wallet"
)

func main() {
	Init()
	RunTx()
}
func Init() {
	// Initialize the configuration center
	config.InitNacos(utils.GetRunTime())
	// 链接数据库
	gorm.NewDB("DbOrder")
	gorm.NewDB("DbWallet")
	gorm.NewDB("DbTx")
	// 创建表
	tx.CreateTable()
	order.CreateTable()
	wallet.CreateTable()
}

type Id struct {
	Id int64 `json:"id" gorm:"column:id"`
}

func RunTx() {
	// 链接redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewClient(&redis.Options{Addr: redisHost, Password: redisPassword, DB: 0})

	// 创建事务
	transaction := tx.Tx{}
	tx.TxHandler.Insert(nil, &transaction)

	// 对uid为1的用户进行加锁
	uid := "1"
	ok, err := redisClient.SetNX("tx_"+uid, transaction.Id, 0).Result()
	if !ok {
		return
	}
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// 开启钱包事务
	walletTx := wallet.Orm().Begin()
	defer walletTx.RollbackUnlessCommitted()
	// now := time.Now().Unix()

	// 开启订单事务
	orderTx := order.Orm().Begin()
	defer orderTx.RollbackUnlessCommitted()
	// 获取当前链接进程Id
	var mysql1 Id
	err = walletTx.Raw("SELECT CONNECTION_ID() as id").Scan(&mysql1).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	wTx1 := wallet.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql1.Id,
	}
	err = wallet.TxRecordHandler.Insert(nil, &wTx1)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}

	// 获取当前链接进程Id
	var mysql2 Id
	err = orderTx.Raw("SELECT CONNECTION_ID() as id").Scan(&mysql2).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	wTx2 := order.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql2.Id,
	}
	err = order.TxRecordHandler.Insert(nil, &wTx2)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	// check
	err = wallet.WalletHandler.Update(walletTx, map[string]interface{}{
		"money": 1,
	}, "id = 1")
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	err = order.OrderHandler.Update(orderTx, map[string]interface{}{
		"status": 1,
	}, "id = 1")
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// preper
	err = wallet.TxRecordHandler.Update(nil, map[string]interface{}{
		"status": 1,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}

	err = order.TxRecordHandler.Update(nil, map[string]interface{}{
		"status": 1,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// commit
	err = wallet.TxRecordHandler.Update(walletTx, map[string]interface{}{
		"status": 2,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}

	err = order.TxRecordHandler.Update(orderTx, map[string]interface{}{
		"status": 2,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}

	orderTx.Commit()
	walletTx.Commit()

}
