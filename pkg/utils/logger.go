package utils

import "github.com/m-mizutani/zlog"

var Logger = zlog.New()

func InitLogger(options ...zlog.Option) {
	Logger = Logger.Clone(options...)
}
