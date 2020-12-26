package cache

func Get(key string) interface{} {
	return database[key]
}

func List() map[string]interface{} {
	return database
}