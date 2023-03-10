package utils

import "github.com/fbsobreira/gotron-sdk/pkg/address"

func ToFixed(address address.Address) [21]byte {
	res := (*[21]byte)(address)
	return *res
}
