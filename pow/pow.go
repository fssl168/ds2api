package pow

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"

	"golang.org/x/crypto/sha3"
)

func SolvePow(ctx context.Context, challenge string, salt string, expireAt int64, difficulty float64) (int64, error) {
	prefix := salt + "_" + fmt.Sprintf("%d", expireAt) + "_"
	diffInt := int64(difficulty)
	if diffInt <= 0 {
		diffInt = 144000
	}
	target := uint64(1) << (64 - uint(math.Ceil(math.Log2(float64(diffInt)))))
	if target == 0 {
		target = 1
	}

	var nonce int64
	buf := make([]byte, len(prefix)+8)
	copy(buf, prefix)
	prefixLen := len(prefix)

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
		binary.LittleEndian.PutUint64(buf[prefixLen:], uint64(nonce))
		h := sha3.Sum256(buf)
		val := binary.LittleEndian.Uint64(h[:8])
		if val < target {
			return nonce, nil
		}
		nonce++
		if nonce&0xfff == 0 {
			bits.LeadingZeros64(val)
		}
	}
}
