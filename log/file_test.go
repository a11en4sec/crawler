package log_test

import (
	"fmt"
	"github.com/a11en4sec/crawler/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	const filePrefix = "test"
	const fileSuffix = ".log"
	const gzipSuffix = ".gz"

	var p, c = log.NewFilePlugin(filePrefix+fileSuffix, zapcore.DebugLevel)

	var logger = log.NewLogger(p)
	var b = make([]byte, 10000)
	var count = 10
	for count > 0 {
		count--
		logger.Info(string(b))
	}

	var err = c.Close()
	require.NoError(t, err)
	// NOTE: 目前Lumberjack的实现，close不会停掉压缩协程
	// 等待Lumberjack压缩日志文件
	time.Sleep(1 * time.Second)

	fs, err := ioutil.ReadDir(".")
	require.NoError(t, err)
	var gzCount, logCount int
	for _, f := range fs {
		var name = f.Name()
		fmt.Printf("name:%s \n", name)
		if strings.HasPrefix(name, filePrefix) {
			if strings.HasSuffix(name, fileSuffix) {
				logCount++
				//assert.NoError(t, os.Remove(f.Name()))
				continue
			}
			if strings.HasSuffix(name, fileSuffix+gzipSuffix) {
				gzCount++
				logCount++
				//assert.NoError(t, os.Remove(f.Name()))
				continue
			}
		}
	}
	//require.Equal(t, 3, logCount)
	//require.Equal(t, 2, gzCount)
}
