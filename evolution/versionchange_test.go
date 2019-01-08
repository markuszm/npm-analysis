package evolution

import (
	"testing"
)

func TestVersionChangesSort(t *testing.T) {
	changes := []VersionChange{
		{
			Version: "0.0.1",
		}, {
			Version: "1.0.1",
		}, {
			Version: "1.3.1",
		}, {
			Version: "0.0.3",
		}, {
			Version: "0.1.10",
		}, {
			Version: "2.0.1",
		},
	}

	SortVersionChange(changes)

	for i, c := range changes {
		switch i {
		case 0:
			assert(c.Version, "0.0.1", i, t)
		case 1:
			assert(c.Version, "0.0.3", i, t)
		case 2:
			assert(c.Version, "0.1.10", i, t)
		case 3:
			assert(c.Version, "1.0.1", i, t)
		case 4:
			assert(c.Version, "1.3.1", i, t)
		case 5:
			assert(c.Version, "2.0.1", i, t)
		}
	}
}

func TestVersionChangesPrereleasesSort(t *testing.T) {
	changes := []VersionChange{
		{
			Version: "0.9.0-beta.10",
		}, {
			Version: "0.9.0-beta.2",
		}, {
			Version: "0.9.0-beta.8",
		}, {
			Version: "1.1.0-beta.2",
		}, {
			Version: "0.1.10",
		}, {
			Version: "2.1.4",
		},
	}

	SortVersionChange(changes)

	for i, c := range changes {
		switch i {
		case 0:
			assert(c.Version, "0.1.10", i, t)
		case 1:
			assert(c.Version, "0.9.0-beta.2", i, t)
		case 2:
			assert(c.Version, "0.9.0-beta.8", i, t)
		case 3:
			assert(c.Version, "0.9.0-beta.10", i, t)
		case 4:
			assert(c.Version, "1.1.0-beta.2", i, t)
		case 5:
			assert(c.Version, "2.1.4", i, t)
		}
	}
}

func assert(actual string, expected string, index int, t *testing.T) {
	if actual != expected {
		t.Errorf("Expected %v but got %v at index %v", expected, actual, index)
	}
}
