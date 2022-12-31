package extensions

import (
	"math/rand"
	"time"
)

func GenerateRandInt(min, max int) int {
	rand.Seed(time.Now().Unix())

	return rand.Intn(max-min) + min
}
