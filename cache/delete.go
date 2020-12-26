package cache

import "github.com/parvez0/redis/responses"

func Delete(key string) (ops *responses.CurdOp) {
	ops = &responses.CurdOp{}
	if database[key] == "" {
		return ops
	}
	mutex.Lock()
	defer mutex.Unlock()
	delete(database, key)
	slaveData := Replicate{
		Action: ReplicateActionDelete,
		Key:    key,
		Data:   nil,
	}
	replicateBucket <- slaveData
	ops.Deleted = 1
	return ops
}
