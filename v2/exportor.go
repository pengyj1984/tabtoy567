package v2

import (
	"container/list"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davyxu/tabtoy/v2/i18n"
	"github.com/davyxu/tabtoy/v2/model"
	"github.com/davyxu/tabtoy/v2/printer"
)

func Run(g *printer.Globals) bool {

	if !g.PreExport() {
		return false
	}

	cachedFile := cacheFile(g)
	fileNames := list.New()
	keys := make([]string, 0, len(cachedFile))
	for k, _ := range cachedFile {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	//log.Infof("==========%s==========", "Begin cached files.")
	//for _, k := range keys {
	//	log.Infof(k)
	//}
	//log.Infof("==========%s==========", "End cached files.")

	// 把标签页搞成文件
	splitedFiles := map[string]*File{}
	for _, k := range keys {
		f := cachedFile[k]
		if strings.ToLower(f.FileName) == "globals.xlsx" {
			splitedFiles[strings.ToLower(f.FileName)] = f
			f.RawFileName = f.FileName
		} else {
			sheets := f.coreFile.Sheets
			for _, sheet := range sheets {
				// 忽略'.'开头的标签页
				if strings.HasPrefix(sheet.Name, ".") {
					continue
				}

				nf := &File{
					valueRepByKey: make(map[valueRepeatData]bool),
					LocalFD:       model.NewFileDescriptor(),
					FileName:      sheet.Name,
				}
				// 根据缓存记录, 给定 Index 值; 如果记录中没有, 就赋新值
				d, ok := g.Data.Tables[nf.FileName]
				if ok {
					nf.LocalFD.SerializeData = d
				} else {
					g.Data.MaxIndex += 1
					nf.LocalFD.SerializeData = model.NewSerializeTableData(g.Data.MaxIndex)
					g.Data.Tables[nf.FileName] = nf.LocalFD.SerializeData
				}

				nf.coreFile = f.coreFile
				nf.RawFileName = f.FileName
				splitedFiles[nf.FileName] = nf
				fileNames.PushBack(nf.FileName)
			}
		}
	}

	fileNames.PushFront("globals.xlsx")
	cachedFile = splitedFiles

	fileObjList := make([]*File, 0)

	log.Infof("==========%s==========", i18n.String(i18n.Run_CollectTypeInfo))

	for fn := fileNames.Front(); fn != nil; fn = fn.Next() {
		//log.Infof("file name = %s", fn.Value)
		file := cachedFile[fn.Value.(string)]
		if file == nil {
			return false
		}

		file.GlobalFD = g.FileDescriptor
		//log.Infof("file descriptor = %s", file.GlobalFD)

		// 电子表格数据导出到Table对象
		if strings.ToLower(file.FileName) == "globals.xlsx" && !file.ExportGlobalType(g) {
			return false
		} else if !file.ExportLocalType(nil, file.LocalFD.SerializeData, g) {
			return false
		}
		// 整合类型信息和数据
		if !g.AddTypes(file.LocalFD) {
			return false
		}
		// 只写入主文件的文件列表(globals 文件不会被加入这个列表中)
		if file.Header != nil {
			fileObjList = append(fileObjList, file)
		}
	}

	// 合并类型
	//for _, in := range g.InputFileList {
	//
	//	inputFile := in.(string)
	//
	//	var mainMergeFile *File
	//
	//	mergeFileList := strings.Split(inputFile, "+")
	//
	//	for index, fileName := range mergeFileList {
	//
	//		file, _ := cachedFile[fileName]
	//
	//		if file == nil {
	//			return false
	//		}
	//
	//		var mergeTarget string
	//		if len(mergeFileList) > 1 {
	//			mergeTarget = "--> " + filepath.Base(mergeFileList[0])
	//		}
	//
	//		log.Infoln(filepath.Base(fileName), mergeTarget)
	//
	//		file.GlobalFD = g.FileDescriptor
	//
	//		// 电子表格数据导出到Table对象
	//		if !file.ExportLocalType(mainMergeFile) {
	//			return false
	//		}
	//
	//		// 主文件才写入全局信息
	//		if index == 0 {
	//
	//			// 整合类型信息和数据
	//			if !g.AddTypes(file.LocalFD) {
	//				return false
	//			}
	//
	//			// 只写入主文件的文件列表
	//			if file.Header != nil {
	//
	//				fileObjList = append(fileObjList, file)
	//			}
	//
	//			mainMergeFile = file
	//		} else {
	//
	//			// 添加自文件
	//			mainMergeFile.mergeList = append(mainMergeFile.mergeList, file)
	//
	//		}
	//
	//	}
	//
	//}

	log.Infof("==========%s==========", i18n.String(i18n.Run_ExportSheetData))

	for _, file := range fileObjList {

		log.Infoln(filepath.Base(file.FileName))

		dataModel := model.NewDataModel()

		tab := model.NewTable()
		tab.LocalFD = file.LocalFD

		// 主表
		if !file.ExportData(dataModel, nil) {
			return false
		}

		// 子表提供数据
		for _, mergeFile := range file.mergeList {

			log.Infoln(filepath.Base(mergeFile.FileName), "--->", filepath.Base(file.FileName))

			// 电子表格数据导出到Table对象
			if !mergeFile.ExportData(dataModel, file.Header) {
				return false
			}
		}

		// 合并所有值到node节点
		if !mergeValues(dataModel, tab, file) {
			return false
		}

		// 整合类型信息和数据
		// 将这个文件的定义添加到 global 的 fields 中
		if !g.AddContent(tab) {
			return false
		}

	}

	// 根据各种导出类型, 调用各导出器导出
	return g.Print()
}
