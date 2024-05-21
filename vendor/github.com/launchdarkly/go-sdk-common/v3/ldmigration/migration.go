package ldmigration

import "fmt"

// ExecutionOrder represents the various execution modes this SDK can operate
// under while performing migration-assisted reads.
type ExecutionOrder string

const (
	// Serial execution ensures the authoritative read will always complete execution before executing the
	// non-authoritative read.
	Serial ExecutionOrder = "serial"
	// Random execution randomly decides if the authoritative read should execute first or second.
	Random ExecutionOrder = "random"
	// Concurrent executes both reads in separate go routines, and waits until both calls have finished before
	// proceeding.
	Concurrent ExecutionOrder = "concurrent"
)

// Operation represents a type of migration operation; namely, read or write.
type Operation string

const (
	// Read denotes a read-related migration operation.
	Read Operation = "read"
	// Write denotes a write-related migration operation.
	Write Operation = "write"
)

// ConsistencyCheck records the results of a consistency check and the ratio at
// which the check was sampled.
//
// For example, a sampling ratio of 10 indicts this consistency check was
// sampled approximately once every ten operations.
type ConsistencyCheck struct {
	consistent    bool
	samplingRatio int
}

// NewConsistencyCheck creates a new consistency check reflecting the provided values.
func NewConsistencyCheck(wasConsistent bool, samplingRatio int) *ConsistencyCheck {
	return &ConsistencyCheck{
		consistent:    wasConsistent,
		samplingRatio: samplingRatio,
	}
}

// Consistent returns whether or not the check returned a consistent result.
func (c ConsistencyCheck) Consistent() bool {
	return c.consistent
}

// SamplingRatio returns the 1 in x sampling ratio used to determine if the consistency check should be run.
func (c ConsistencyCheck) SamplingRatio() int {
	return c.samplingRatio
}

// Origin represents the source of origin for a migration-related operation.
type Origin string

const (
	// Old represents the technology source we are migrating away from.
	Old Origin = "old"
	// New represents the technology source we are migrating towards.
	New Origin = "new"
)

// Stage denotes one of six possible stages a technology migration could be a
// part of, progressing through the following order.
//
// Off -> DualWrite -> Shadow -> Live -> RampDown -> Complete
type Stage string

const (
	// Off - migration hasn't started, "old" is authoritative for reads and writes
	Off Stage = "off"

	// DualWrite - write to both "old" and "new", "old" is authoritative for reads
	DualWrite Stage = "dualwrite"

	// Shadow - both "new" and "old" versions run with a preference for "old"
	Shadow Stage = "shadow"

	// Live - both "new" and "old" versions run with a preference for "new"
	Live Stage = "live"

	// RampDown - only read from "new", write to "old" and "new"
	RampDown Stage = "rampdown"

	// Complete - migration is done
	Complete Stage = "complete"
)

// ParseStage parses a MigrationStage from a string, or returns an error if the stage is unrecognized.
func ParseStage(val string) (Stage, error) {
	switch val {
	case "off":
		return Off, nil
	case "dualwrite":
		return DualWrite, nil
	case "shadow":
		return Shadow, nil
	case "live":
		return Live, nil
	case "rampdown":
		return RampDown, nil
	case "complete":
		return Complete, nil
	default:
		return Off, fmt.Errorf("invalid stage %s provided", val)
	}
}
