# cache

Cache using Go 1.18 generics.

`cache` is an in-memory key:value store/cache similar to memcached that is
suitable for applications running on a single machine.

```go
import (
	"fmt"
	"github.com/alaingilbert/cache"
)

func main() {
	// Create a cache of "string" with a default expiration time of 5 minutes
	// and which purges expired items every 10 minutes (DefaultCleanupInterval)
	// The cleanup interval can be changed for a custom value using "CleanupInterval"
	// `cache.New[string](5*time.Minute, cache.CleanupInterval(time.Hour))`
	// The cleanup thread can be terminated using a context.Context
	// `cache.New[string](5*time.Minute, cache.WithContext(ctx))`
	c := cache.New[string](5*time.Minute)
	
	// Set the value of the key "key1" to "val1", with the default expiration time (5min)
	c.Set("key1", "val1")

	// Set the value of the key "key2" to "val2", with a 1sec expiration time
	c.Set("key2", "val2", cache.ExpireIn(time.Second))
	
	// Set the value of the key "key3" to "val3", which will expire on Jan 01 2100
	c.Set("key3", "val3", cache.ExpireAt(time.Date(2100, 1, 1, 0, 0, 0, 0, time.Local)))
	
	// Set the value of the key "key4" to "val4", with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("key4")
	c.Set("key4", "val4", cache.NoExpire)
	
	// Return either or not "key1" is in the cache and not expired
	found := c.Has("key1")

	// Get the string associated with the key "key1" from the cache
	if value, found := c.Get("key1"); found {
		fmt.Println(value)
	}

	// Delete "key1" from the cache
	c.Delete("key1")
	
	// Can also use a "Set" (data structure) for cache
	s := cache.NewSet[string](5*time.Minute)
	s.Set("key1")
	if s.Has("key1") {
		fmt.Println("found key1")
	}
	s.Delete("key1")
}
```