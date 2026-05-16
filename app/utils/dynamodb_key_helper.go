package utils

import (
	"time"

	"github.com/google/uuid"
)

func GetPartitionKey(pkPrefix string, id string) string {
	return pkPrefix + "#" + id
}

func GetSortingKey(skPrefix string, date time.Time, id string) string {
	return skPrefix + "#" + date.UTC().Format("2006-01-02") + "#" + id
}

func GetUUID() string {
	return uuid.NewString()
}
