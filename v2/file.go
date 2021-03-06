package v2

import (
	"fmt"
	"github.com/davyxu/tabtoy/util"
	"github.com/davyxu/tabtoy/v2/printer"
	"strings"

	"github.com/davyxu/tabtoy/v2/i18n"
	"github.com/davyxu/tabtoy/v2/model"
	"github.com/tealeg/xlsx"
)

// 检查单元格值重复结构
type valueRepeatData struct {
	fd    *model.FieldDescriptor
	value string
}

// 1个电子表格文件
type File struct {
	LocalFD     *model.FileDescriptor // 本文件的类型描述表
	GlobalFD    *model.FileDescriptor // 全局的类型描述表
	FileName    string
	RawFileName string
	coreFile    *xlsx.File

	dataSheets  []*DataSheet
	Header      *DataHeader
	dataHeaders []*DataHeader

	valueRepByKey map[valueRepeatData]bool // 检查单元格值重复map

	mergeList []*File
}

func (self *File) GlobalFileDesc() *model.FileDescriptor {
	return self.GlobalFD

}

// 处理表(标签页)的类型信息
func (self *File) ExportLocalType(mainFile *File, tableData *model.SerializeTableData, g *printer.Globals) bool {

	//var sheetCount int

	//var typeSheet *TypeSheet
	//// 解析类型表
	//// 这一步只有处理globals表的时候需要, 其他时候可以考虑去掉...... 想一想怎么去合理
	//for _, rawSheet := range self.coreFile.Sheets {
	//
	//	if isTypeSheet(rawSheet.Name) {
	//		if sheetCount > 0 {
	//			log.Errorf("%s", i18n.String(i18n.File_TypeSheetKeepSingleton))
	//			return false
	//		}
	//
	//		typeSheet = newTypeSheet(NewSheet(self, rawSheet))
	//
	//		// 从cell添加类型
	//		if !typeSheet.Parse(self.LocalFD, self.GlobalFD) {
	//			return false
	//		}
	//
	//		sheetCount++
	//
	//	}
	//}

	// 2021.6.16 -- 类型信息统一放在 globals.xlsx 中, 就不要求每个文件都有这个页签了
	//if typeSheet == nil {
	//	log.Errorf("%s", i18n.String(i18n.File_TypeSheetNotFound))
	//	return false
	//}

	dataSheetName := ""
	// 解析表头
	// globals 这里不会处理
	for _, rawSheet := range self.coreFile.Sheets {

		// 因为加入了分标签页的机制, 这里需要判断是不是我们想要处理的表。即判断标签页名和记录的fileName是否一致
		if !isTypeSheet(rawSheet.Name) && rawSheet.Name == self.FileName {
			dSheet := newDataSheet(NewSheet(self, rawSheet))

			if !dSheet.Valid() {
				continue
			}

			dataSheetName = rawSheet.Name
			//if typeSheet == nil {
			//	self.LocalFD.Pragma.KVPair.SetString("TableName", dataSheetName)
			//	self.LocalFD.Pragma.KVPair.SetString("Package", "table")
			//	self.LocalFD.Pragma.KVPair.SetString("CSClassHeader", "[System.Serializable]")
			//}
			self.LocalFD.Pragma.KVPair.SetString("TableName", dataSheetName)
			self.LocalFD.Pragma.KVPair.SetString("Package", "table")
			self.LocalFD.Pragma.KVPair.SetString("CSClassHeader", "[System.Serializable]")
			//self.LocalFD.Pragma.KVPair.Parse("OutputTag:['.lua', '.cs', '.json', '.go']")
			log.Infof("            %s", rawSheet.Name)

			dataHeader := newDataHeadSheet()

			// 检查引导头
			// 初始化表中的数据结构
			if !dataHeader.ParseProtoField(len(self.dataSheets), dSheet.Sheet, self.LocalFD, self.GlobalFD, tableData) {
				return false
			}

			if mainFile != nil {

				if fieldName, ok := dataHeader.AsymmetricEqual(mainFile.Header); !ok {
					log.Errorf("%s main: %s child: %s field: %s", i18n.String(i18n.DataHeader_NotMatchInMultiTableMode), mainFile.FileName, self.FileName, fieldName)
					return false
				}

			}

			if self.Header == nil {
				self.Header = dataHeader
			}

			self.dataHeaders = append(self.dataHeaders, dataHeader)
			self.dataSheets = append(self.dataSheets, dSheet)
			// 只处理我们要得标签页, 其他的不用管了
			break
		}
	}

	// File描述符的名字必须放在类型里, 因为这里始终会被调用, 但是如果数据表缺失, 是不会更新Name的
	self.LocalFD.Name = self.LocalFD.Pragma.GetString("TableName")
	self.LocalFD.SerializeData = tableData
	tag, ok := g.OutputTags[self.LocalFD.Name]
	if ok{
		self.LocalFD.Pragma.KVPair.Parse(fmt.Sprintf("OutputTag:%s", tag))
	}

	return true
}

func (self *File) ExportGlobalType(g *printer.Globals) bool {
	var sheetCount int

	var typeSheet *TypeSheet
	// 解析类型表
	// 这一步只有处理globals表的时候需要, 其他文件不管
	for _, rawSheet := range self.coreFile.Sheets {

		if isTypeSheet(rawSheet.Name) {
			if sheetCount > 0 {
				log.Errorf("%s", i18n.String(i18n.File_TypeSheetKeepSingleton))
				return false
			}

			typeSheet = newTypeSheet(NewSheet(self, rawSheet))

			// 从cell添加类型
			if !typeSheet.Parse(self.LocalFD, self.GlobalFD, g.Data) {
				return false
			}

			sheetCount++

		} else if isOutputSheet(rawSheet.Name){
			// 2021.8.1 增加 OutputTag 标签页
			for row := 0; row < rawSheet.MaxRow; row++{
				table := strings.TrimSpace(rawSheet.Row(row).Cells[0].String())
				if len(table) == 0{
					// 遇到空单元格就直接忽略后面的
					break
				}

				if table[0] == '#'{
					// 注释行
					continue
				}

				tag := strings.TrimSpace(rawSheet.Row(row).Cells[1].String())
				if len(tag) == 0 || tag[0] == '#'{
					// 无效内容, 直接跳过
					continue
				}

				_, ok := g.OutputTags[table]
				if ok{
					log.Infoln("输出列表中已经包含 %s 表设置, 请勿重复设置", table)
					continue
				}

				g.OutputTags[table] = tag
			}
		}
	}

	// 2021.6.16 -- 类型信息统一放在 globals.xlsx 中, 就不要求每个文件都有这个页签了
	if typeSheet == nil {
		log.Errorf("%s", i18n.String(i18n.File_TypeSheetNotFound))
		return false
	}

	// File描述符的名字必须放在类型里, 因为这里始终会被调用, 但是如果数据表缺失, 是不会更新Name的
	self.LocalFD.Name = self.LocalFD.Pragma.GetString("TableName")

	return true
}

func (self *File) IsVertical() bool {
	return self.LocalFD.Pragma.GetBool("Vertical")
}

func (self *File) ExportData(dataModel *model.DataModel, parentHeader *DataHeader) bool {

	for index, d := range self.dataSheets {

		log.Infof("            %s", d.Name)

		// 多个sheet时, 使用和多文件一样的父级
		if parentHeader == nil && len(self.dataHeaders) > 1 {
			parentHeader = self.dataHeaders[0]
		}

		if !d.Export(self, dataModel, self.dataHeaders[index], parentHeader) {
			return false
		}
	}

	return true

}

func (self *File) CheckValueRepeat(fd *model.FieldDescriptor, value string) bool {

	key := valueRepeatData{
		fd:    fd,
		value: value,
	}

	if _, ok := self.valueRepByKey[key]; ok {
		return false
	}

	self.valueRepByKey[key] = true

	return true
}

func isTypeSheet(name string) bool {
	return strings.TrimSpace(name) == model.TypeSheetName
}

func isOutputSheet(name string) bool{
	return strings.TrimSpace(name) == model.OutputSheetName
}

func NewFile(filename string, cacheDir string) (f *File, fromCache bool) {

	self := &File{
		valueRepByKey: make(map[valueRepeatData]bool),
		LocalFD:       model.NewFileDescriptor(),
		FileName:      filename,
	}

	if cacheDir != "" {
		cache := util.NewTableCache(filename, cacheDir)

		if err := cache.Open(); err != nil {
			log.Errorf("%s, %v", i18n.String(i18n.System_OpenReadXlsxFailed), err.Error())
			return nil, false
		}

		if cfile, err := cache.Load(); err != nil {
			log.Errorln(err.Error())
			log.Errorf("%s, %v", i18n.String(i18n.System_OpenReadXlsxFailed), err.Error())
			return nil, false
		} else {
			self.coreFile = cfile

			if !cache.UseCache() {
				cache.Save()
			}

			return self, cache.UseCache()
		}
	}

	var err error
	self.coreFile, err = xlsx.OpenFile(filename)
	if err != nil {
		log.Errorf("%s, %v", i18n.String(i18n.System_OpenReadXlsxFailed), err.Error())
		return nil, false
	}

	return self, false
}
