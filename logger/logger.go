package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

type LogParams struct {
	LogMode   string `default:""` // this only works with strings
	LogFile   string `default:"file_manager.log"`
	LogPrefix string `default:""`
}

func InitLogger(params *LogParams) *log.Logger {
	var (
		out io.Writer = ioutil.Discard
		err error
	)
	typ := reflect.TypeOf(*params)

	if params.LogMode == "" {
		f, _ := typ.FieldByName("LogMode")
		params.LogMode = f.Tag.Get("default")
	}

	if params.LogFile == "" {
		f, _ := typ.FieldByName("LogFile")
		params.LogFile = f.Tag.Get("default")
	}

	if params.LogPrefix == "" {
		f, _ := typ.FieldByName("LogPrefix")
		params.LogPrefix = f.Tag.Get("default")
	}
	switch params.LogMode {
	case "file":
		out, err = os.Create(params.LogFile)
		if nil != err {
			panic(err.Error())
		}
	case "screen":
		out = os.Stdout

	default:
		out = ioutil.Discard
	}

	return log.New(out, params.LogPrefix, log.Lshortfile)
}
