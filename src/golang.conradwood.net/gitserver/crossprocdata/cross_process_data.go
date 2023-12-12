package crossprocdata

import (
	"golang.conradwood.net/go-easyops/cache"
	"time"
)

/*
purpose: keep some short-lived data around that can be passed with a single reference.
for example an environment variable
*/

var (
	local_cache = cache.New("localdata", time.Duration(90)*time.Second, 100)
)

// safe to add more stuff here
type LocalData struct {
	HTTPRequest interface{}
}

// might be nil
func GetLocalData(key string) *LocalData {
	obj := local_cache.Get(key)
	if obj == nil {
		return nil
	}
	return obj.(*LocalData)
}

func SaveLocalData(key string, ld *LocalData) {
	local_cache.Put(key, ld)
}



