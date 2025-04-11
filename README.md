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
	c := cache.New[string](5*time.Minute)
	c.Set("key1", "val1")
	c.Set("key2", "val2")
	found := c.Has("key1")
	value, found := c.Get("key1")
	if found {
		fmt.Println(value)
	}
	c.Delete("key1")
}
```