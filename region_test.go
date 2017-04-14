package types

import (
	"testing"
)

// TestRegionUnmarshal_NoCompare tries to unmarshal all json test data for this type, but doesn't compare its response to a known correct output.
func TestRegionUnmarshal_NoCompare(t *testing.T) {
	// Get the input
	data := testData["region"].known
	if len(data) == 0 {
		t.Skip("No data to test")
	}

	// Create the run function generator, allowing us to run it in parallel
	rgen := func(in []byte) func(t *testing.T) {
		return func(t *testing.T) {
			var r = &Region{}

			err := r.UnmarshalJSON(in)
			if err != nil {
				t.Errorf("Error while unmarshalling: %v", err)
			}
		}
	}

	// For each of them, let's run a subtest
	for name, datum := range data {
		// Get the run function
		rfunc := rgen(datum.raw)

		// Run !
		t.Run(name, rfunc)
	}
}
