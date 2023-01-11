package sqlstorage

import (
	"encoding/json"
	"errors"

	"github.com/a11en4sec/crawler/spider"

	"github.com/a11en4sec/crawler/engine"
	"github.com/a11en4sec/crawler/sqldb"
	"go.uber.org/zap"
)

type SQLStorage struct {
	dataDocker []*spider.DataCell //分批输出结果缓存
	//columnNames []sqldb.Field      //标题字段
	db    sqldb.DBer //接口
	Table map[string]struct{}
	options
}

func New(opts ...Option) (*SQLStorage, error) {
	options := defaultoptions
	// 每个函数，一次作用于默认的选项(不停被修改、增强)
	for _, opt := range opts {
		opt(&options)
	}

	s := &SQLStorage{}
	s.options = options
	s.Table = make(map[string]struct{})

	var err error
	// s.db 是接口sqldb.DBer, sqldb.New返回是*Sqldb,实现了接口DBer,所以可以赋值
	s.db, err = sqldb.New(sqldb.WithConnURL(s.sqlURL), sqldb.WithLogger(s.logger))

	if err != nil {
		return nil, err
	}

	return s, nil
}

// Save 实现storage
func (s *SQLStorage) Save(dataCells ...*spider.DataCell) error {
	for _, cell := range dataCells {
		name := cell.GetTableName()
		if _, ok := s.Table[name]; !ok {
			// 没有表的话，创建表
			// 1. 获取列
			columnNames := getFields(cell)
			// 2. 创建
			err := s.db.CreateTable(sqldb.TableData{
				TableName:   name,
				ColumnNames: columnNames,
				Args:        nil,
				DataCount:   0,
				AutoKey:     true,
			})

			if err != nil {
				s.logger.Error("create table failed", zap.Error(err))

				continue
			}
			// 表已经创建了，此处赋一个空结构体
			s.Table[name] = struct{}{}
		}

		// 已满足batch的条件，刷盘
		if len(s.dataDocker) >= s.BatchCount {
			err := s.Flush()
			if err != nil {
				s.logger.Error("flush into disk failed", zap.Error(err))
			}
		}
		// 暂时存储在dataDocker仓库缓存中
		s.dataDocker = append(s.dataDocker, cell)
	}

	return nil
}

func getFields(cell *spider.DataCell) []sqldb.Field {
	taskName := cell.Data["Task"].(string)
	ruleName := cell.Data["Rule"].(string)
	fields := engine.GetFields(taskName, ruleName)

	var columnNames []sqldb.Field
	for _, field := range fields {
		columnNames = append(columnNames, sqldb.Field{
			Title: field,
			Type:  "MEDIUMTEXT", // sql中用于存放字符串的类型
		})
	}

	columnNames = append(columnNames,
		sqldb.Field{Title: "URL", Type: "VARCHAR(255)"},
		sqldb.Field{Title: "Time", Type: "VARCHAR(255)"},
	)

	return columnNames
}

func (s *SQLStorage) Flush() error {
	if len(s.dataDocker) == 0 {
		return nil
	}

	defer func() {
		s.dataDocker = nil
	}()

	args := make([]interface{}, 0)
	var ruleName string
	var taskName string
	var ok bool
	// dataDocker满了
	for _, datacell := range s.dataDocker {
		if ruleName, ok = datacell.Data["Rule"].(string); !ok {
			return errors.New("no rule field")
		}

		if taskName, ok = datacell.Data["Task"].(string); !ok {
			return errors.New("no task field")
		}
		// 用于获取当前数据的表字段与字段类型
		fields := engine.GetFields(taskName, ruleName)

		data := datacell.Data["Data"].(map[string]interface{})

		var value []string

		for _, field := range fields {
			v := data[field]
			switch v := v.(type) {
			case nil:
				value = append(value, "")
			case string:
				value = append(value, v)
			default:
				// todo:为什么序列化
				j, err := json.Marshal(v)
				if err != nil {
					value = append(value, "")
				} else {
					value = append(value, string(j))
				}
			}
		}

		//value = append(value, datacell.Data["URL"].(string), datacell.Data["Time"].(string))
		if v, ok := datacell.Data["URL"].(string); ok {
			value = append(value, v)
		}
		if v, ok := datacell.Data["Time"].(string); ok {
			value = append(value, v)
		}

		for _, v := range value {
			args = append(args, v)
		}
	}

	return s.db.Insert(sqldb.TableData{
		TableName:   s.dataDocker[0].GetTableName(),
		ColumnNames: getFields(s.dataDocker[0]),
		Args:        args,              // 插入的值
		DataCount:   len(s.dataDocker), // 满了才插入，一次插入的数量就是dataDocker的大小
	})
}
