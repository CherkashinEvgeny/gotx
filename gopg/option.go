package sql

import (
	"github.com/CherkashinEvgeny/gotx/base"
)

type Option = base.Option

const propagationKey = "propagation"

func WithPropagation(propagation Propagation) Option {
	return Option{
		Key:   propagationKey,
		Value: propagation,
	}
}

func getPropagation(options base.Options) (propagation Propagation) {
	propagation, _ = options.Value(propagationKey).(Propagation)
	return propagation
}

type Propagation int

const (
	Never Propagation = iota - 2
	Supports
	Required
	Nested
	Mandatory
)

const isolationLevelKey = "isolation"

func WithIsolationLevel(level Isolation) (option Option) {
	return base.Option{
		Key:   isolationLevelKey,
		Value: level,
	}
}

func getIsolationLevel(options base.Options) (level Isolation) {
	level, _ = options.Value(isolationLevelKey).(Isolation)
	return level
}

type Isolation int

func (i Isolation) String() (str string) {
	switch i {
	case ReadCommitted:
		return "ReadCommitted"
	case RepeatableRead:
		return "RepeatableRead"
	case Serializable:
		return "Serializable"
	default:
		return ""
	}
}

const (
	ReadUncommitted Isolation = 0
	ReadCommitted   Isolation = 0
	RepeatableRead  Isolation = 1
	Serializable    Isolation = 2
)
