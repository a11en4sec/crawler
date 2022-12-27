package storage

type OutputData struct {
	Data map[string]interface{}
}

type Store interface {
	Save(datas ...OutputData) error
}
