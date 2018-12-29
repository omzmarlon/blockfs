package chain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/omzmarlon/blockfs/bclib"
)

// ComputeHashString - static function to compute hash of given input
func ComputeHashString(prevHash string, minerID string, nonce uint32, ops []bclib.ROp) string {
	md5Hasher := md5.New()
	md5Hasher.Write([]byte(prevHash + minerID + fmt.Sprint(nonce) + fmt.Sprintf("%v", ops)))
	return hex.EncodeToString(md5Hasher.Sum(nil))
}

func targetZeros(difficulty uint8) string {
	target := ""

	for i := 0; i < int(difficulty); i++ {
		target += "0"
	}
	return target
}

func CheckDifficultyReached(hash string, ops []bclib.ROp, opD int, noopD int) bool {
	var difficulty uint8
	if ops == nil || len(ops) == 0 {
		difficulty = uint8(noopD)
	} else {
		difficulty = uint8(opD)
	}
	if hash[len(hash)-int(difficulty):] == targetZeros(difficulty) {
		return true
	}
	return false
}
