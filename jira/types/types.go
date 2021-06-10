package types

import "strconv"

type AccID int64

func (a AccID) String() string {
	return strconv.FormatUint(uint64(a), 10)
}
