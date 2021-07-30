package model

import (
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("model")

const TypeSheetName = "@Types"
// 2021.7.30 增加一个输出标签页, 标记每个表输出哪些格式. 默认是全格式输出
const OutputSheetName = "@Outputs"
