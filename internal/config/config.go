package config

import (
	"sync"
	"sync/atomic"
)

var initOnce sync.Once
var gcf atomic.Value
