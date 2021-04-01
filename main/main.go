package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/redis.v5"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cast"
	gormx "github.com/sunmi-OS/gocore/gorm"
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
	gormx.NewDB("DbOrder")
	gormx.NewDB("DbWallet")
	gormx.NewDB("DbTx")
	// 创建表
	tx.CreateTable()
	order.CreateTable()
	wallet.CreateTable()
}

type Id struct {
	Id int64 `json:"id" gorm:"column:id"`
}

type TrxId struct {
	TrxId int64 `json:"trx_id" gorm:"column:trx_id"`
}

func RunTx() {
	// 链接redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewClient(&redis.Options{Addr: redisHost, Password: redisPassword, DB: 0})

	// 创建事务
	transaction := tx.Tx{Uid: 1}
	tx.TxHandler.Insert(nil, &transaction)

	// 对uid为1的用户进行加锁
	uid := cast.ToString(transaction.Uid)
	ok, err := redisClient.SetNX("tx_"+uid, transaction.Id, 0).Result()
	if !ok {
		// return
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
	err = walletTx.Raw("SELECT TRX_ID as id FROM INFORMATION_SCHEMA.INNODB_TRX  WHERE TRX_MYSQL_THREAD_ID = CONNECTION_ID()").Scan(&mysql1).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	contentByte, err := json.Marshal(map[string]interface{}{
		"money": "money - 1",
	})
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	wTx1 := wallet.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql1.Id,
		Content: string(contentByte),
	}
	err = wallet.TxRecordHandler.Insert(nil, &wTx1)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}

	// 获取当前链接进程Id
	var mysql2 Id
	err = orderTx.Raw("SELECT TRX_ID as id FROM INFORMATION_SCHEMA.INNODB_TRX  WHERE TRX_MYSQL_THREAD_ID = CONNECTION_ID()").Scan(&mysql2).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	contentByte, err = json.Marshal(map[string]interface{}{
		"status": 1,
	})
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}

	wTx2 := order.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql2.Id,
		Content: string(contentByte),
	}
	err = order.TxRecordHandler.Insert(nil, &wTx2)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	// check
	money := "money - 1"
	err = wallet.WalletHandler.Update(walletTx, map[string]interface{}{
		"money": gorm.Expr(money),
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
	tx.TxHandler.Update(nil, map[string]interface{}{
		"status": 1,
	}, "id = ?", transaction.Id)
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
	time.Sleep(time.Hour)

	orderTx.Commit()
	walletTx.Commit()
	redisClient.Del("tx_" + uid)
	tx.TxHandler.Update(nil, map[string]interface{}{
		"status": 2,
	}, "id = ?", transaction.Id)

}

func TxCron() {
	// 获取当前所有的事物
	var walletTrxId []*TrxId
	err := wallet.Orm().Raw("select trx_id from information_schema.INNODB_TRX;").Scan(&walletTrxId).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	walletTrxMap := make(map[int64]bool)
	for k1 := range walletTrxId {
		walletTrxMap[walletTrxId[k1].TrxId] = true
	}

	var orderTrxId []*TrxId
	err = order.Orm().Raw("select trx_id from information_schema.INNODB_TRX;").Scan(&orderTrxId).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	orderTrxMap := make(map[int64]bool)
	for k1 := range orderTrxId {
		orderTrxMap[orderTrxId[k1].TrxId] = true
	}
	unPreperList, _ := tx.TxHandler.GetList("status = 0")
	for k1 := range unPreperList {
		obj := unPreperList[k1]
		fmt.Printf("%#v\n", obj)
	}
	preperList, _ := tx.TxHandler.GetList("status = 1")
	for k1 := range preperList {
		obj := preperList[k1]
		fmt.Printf("%#v\n", obj)
	}
}
