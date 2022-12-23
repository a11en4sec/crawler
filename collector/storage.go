package collector

type Storage interface {
	Save(datas ...*DataCell) error
}

// DataCell 是对数据的抽象
type DataCell struct {
	Data map[string]interface{}
	// Data["Task"] 当前的任务名
	// Data["Rule"] 当前的规则名
	// Data["Url"] 当前的网址
	// Data["Time"]
	// Data["Data"] 当前核心的数据，即当前书籍的详细信息

}

func (d *DataCell) GetTableName() string {
	return d.Data["Task"].(string)
}

func (d *DataCell) GetTaskName() string {
	return d.Data["Task"].(string)
}
