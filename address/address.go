package address

import (
	"demo/u"
)

var ServiceAddresses = map[u.Service]string{
	"ReviewCM":  "localhost:8083",
	"StorageCM": "localhost:8084",
}
