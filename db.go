package main

type DB struct {
	data map[string]string
}

func (d *DB) Init() {
	d.data = make(map[string]string)
}

func (d *DB) Get(key string) (data string, ok bool) {
	data, ok = d.data[key]
	return
}

func (d *DB) Set(key string, value string) error {
	d.data[key] = value
	return nil
}
