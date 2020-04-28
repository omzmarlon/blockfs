package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/omzmarlon/blockfs/pkg/domain"
)

// ComputeBlockHash - helper function compute the hash of a block
func ComputeBlockHash(block domain.Block) string {
	md5Hasher := md5.New()
	md5Hasher.Write([]byte(block.PrevHash + block.MinerID + fmt.Sprint(block.Nonce) + fmt.Sprintf("%v", block.Ops)))
	return hex.EncodeToString(md5Hasher.Sum(nil))
}

// IsDifficultyReached - check if the hash reaches desired difficulty
func IsDifficultyReached(hash string, difficulty int) bool {
	if hash[len(hash)-int(difficulty):] == targetZeros(difficulty) {
		return true
	}
	return false
}

// RandomNonce - generates random nonce
func RandomNonce() uint32 {
	source := rand.NewSource(time.Now().UnixNano())
	ran := rand.New(source)
	return ran.Uint32()
}

func targetZeros(difficulty int) string {
	target := ""

	for i := 0; i < int(difficulty); i++ {
		target += "0"
	}
	return target
}
