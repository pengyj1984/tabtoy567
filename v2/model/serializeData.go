/*
	2021.06.22
*/
package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type SerializeTableData struct {
	MaxOrder    int32            `json:"MaxOrder"`    // 当前最大序号, 用于给新的值用
	FieldOrders map[string]int32 `json:"FieldOrders"` // 记录 字段名-序号
	Index       int32            `json:"Index"`       // 下标
}

type SerializeData struct {
	MaxIndex int32                          `json:"MaxIndex"` // 记录当前最大下标值, 用于给新的表用
	Tables   map[string]*SerializeTableData `json:"Tables"`   // 表信息
}

func NewSerializeData() *SerializeData {
	self := &SerializeData{
		MaxIndex: 0,
	}

	fileName := "cache.json"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		log.Errorln("File not exists error")
		self.Tables = make(map[string]*SerializeTableData)
	} else {
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Errorln("Read error, err = ", err)
			self.Tables = make(map[string]*SerializeTableData)
		} else {
			err := json.Unmarshal(data, self)
			if err != nil {
				log.Errorln("Unmarshal error, err = ", err)
				self.Tables = make(map[string]*SerializeTableData)
			}
		}
	}

	return self
}

func (self *SerializeData) WriteSerializeData() error {
	fileName := "cache.json"
	_, err := os.Stat(fileName)
	if !(os.IsNotExist(err)) {
		os.Remove(fileName)
	}

	bytes, err := json.Marshal(self)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, bytes, os.ModeAppend)
	if err != nil {
		return nil
	}

	return nil
}

func NewSerializeTableData(index int32) *SerializeTableData {
	self := &SerializeTableData{
		MaxOrder:    0,
		Index:       index,
		FieldOrders: make(map[string]int32),
	}

	return self
}
