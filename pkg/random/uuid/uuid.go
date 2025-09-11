package uuid

import (
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

func getStandardUUID() string {
	return uuid.New().String()
}

func GetUUID() string {
	return strings.ReplaceAll(getStandardUUID(), "-", "")
}

// nolint: gosec
func GetUUIDWithSeed(seed int64) string {
	r := rand.New(rand.NewSource(seed))
	uu, _ := uuid.NewRandomFromReader(r)
	return strings.ReplaceAll(uu.String(), "-", "")
}
