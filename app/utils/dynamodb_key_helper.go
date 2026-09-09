package utils

import (
	"time"

	"github.com/google/uuid"
)

func GetPartitionKey(pkPrefix string, id string) string {
	return pkPrefix + "#" + id
}

func GetPartitionKeyWithDate(pkPrefix string, date time.Time, id string) string {
	return pkPrefix + "#" + date.UTC().Format("2006-01-02") + "#" + id
}

func GetPartitionKeySubModel(pkPrefix string, pkID string, pkSubModel string, pkSubModelID string) string {
	return pkPrefix + "#" + pkID + "#" + pkSubModel + "#" + pkSubModelID
}

func GetUUID() string {
	return uuid.NewString()
}
