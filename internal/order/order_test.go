package order

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsOrderNumberCorrect(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		isValid bool
	}{
		{
			name:    "Random string",
			val:     "iLoveGo",
			isValid: false,
		},
		{
			name:    "String with number prefix",
			val:     "1337Abakada",
			isValid: false,
		},
		{
			name:    "Random string #2",
			val:     "qwertyuiopasdfgh",
			isValid: false,
		},
		{
			name:    "Invalid number",
			val:     "1233444455556666",
			isValid: false,
		},
		{
			name:    "Valid number #1",
			val:     "12345678903",
			isValid: true,
		},
		{
			name:    "Valid number #2",
			val:     "9278923470",
			isValid: true,
		},
		{
			name:    "Valid number #3",
			val:     "346436439",
			isValid: true,
		},
		{
			name:    "Valid number #4",
			val:     "2377225624",
			isValid: true,
		},
		{
			name:    "Invalid number #2",
			val:     "2377225626",
			isValid: false,
		},
		{
			name:    "Valid number #3",
			val:     "9278923473",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, IsOrderNumberCorrect(tt.val))
		})
	}
}
