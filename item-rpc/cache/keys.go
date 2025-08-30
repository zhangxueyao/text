package cachex

import "fmt"

func ItemKey(id int64) string { return fmt.Sprintf("item-api:%d", id) }
