package types

import (
	"github.com/erigontech/erigon/zk/datastream/proto/datastream"
)

type Debug struct {
	Message string
}

func ProcessDebug(debug *datastream.Debug) Debug {
	result := Debug{}

	if debug != nil {
		result.Message = debug.Message
	}

	return result
}
