package service

import "errors"

// ErrNotFound 表示在数据库或其它存储中没找到对应记录
var ErrNotFound = errors.New("resource not found")
