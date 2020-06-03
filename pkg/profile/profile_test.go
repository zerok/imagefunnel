package profile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfileMatching(t *testing.T) {
	// An empty profile should match nothing.
	t.Run("empty-profile", func(t *testing.T) {
		p := Profile{
			Source: Source{},
		}

		require.False(t, p.Matches("something.large.jpg"))
	})

	t.Run("positive-include", func(t *testing.T) {
		p := Profile{
			Source: Source{
				Include: []string{".*\\.jpg"},
			},
		}

		require.True(t, p.Matches("something.large.jpg"))
	})

	t.Run("includes-and-excluded", func(t *testing.T) {
		p := Profile{
			Source: Source{
				Include: []string{".*\\.jpg"},
				Exclude: []string{".*large.*"},
			},
		}

		require.False(t, p.Matches("something.large.jpg"))
	})
}
