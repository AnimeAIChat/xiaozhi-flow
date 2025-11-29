package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite3", "./data/xiaozhi.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// 查询所有记录
	rows, err := db.Query("SELECT id, key, value FROM config_records")
	if err != nil {
		log.Fatal("Failed to query records:", err)
	}
	defer rows.Close()

	var (
		id    int
		key   string
		value string
	)

	fmt.Println("开始修复配置记录的JSON格式...")

	for rows.Next() {
		if err := rows.Scan(&id, &key, &value); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// 将值转换为JSON
		var jsonValue []byte
		var err error

		// 尝试解析现有的值
		var parsedValue interface{}
		if err := json.Unmarshal([]byte(value), &parsedValue); err == nil {
			// 如果已经是有效的JSON，重新序列化确保格式正确
			jsonValue, err = json.Marshal(parsedValue)
			if err != nil {
				log.Printf("Failed to marshal JSON for key %s: %v", key, err)
				continue
			}
		} else {
			// 如果不是有效的JSON，将原始值作为字符串处理
			jsonValue, err = json.Marshal(value)
			if err != nil {
				log.Printf("Failed to marshal string for key %s: %v", key, err)
				continue
			}
		}

		// 更新数据库记录
		_, err = db.Exec("UPDATE config_records SET value = ? WHERE id = ?", string(jsonValue), id)
		if err != nil {
			log.Printf("Failed to update record %d (%s): %v", id, key, err)
			continue
		}

		fmt.Printf("修复记录 %d: %s -> %s\n", id, key, string(jsonValue))
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}

	fmt.Println("配置记录JSON格式修复完成!")
}