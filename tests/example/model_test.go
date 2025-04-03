package example

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddInlinedBodyRequest(t *testing.T) {
	testCases := []struct {
		Name     string
		Raw      string
		Expected AddInlinedBodyRequest
		ErrorMsg string
	}{
		{
			Name:     "minimal",
			Raw:      `{"f07": "test", "f10": ["a"]}`,
			Expected: AddInlinedBodyRequest{F04: 1, F07: "test", F08: AddInlinedBodyRequestF08ValueA, F10: []string{"a"}},
		},
		{
			Name: "maximal",
			Raw: `{"f01": true, "f02": 2, "f03": 3, "f04": 4, "f05": "2025-01-01T12:00:00Z", "f06": "b043c679-354a-4170-a061-dfe2271b3c77", 
"f07": "f7Test", "f08": "valueB", "f09": "summer", "f10": ["a", "b"], "f11": [4, 5], "f12": ["fall", "spring"]}`,
			Expected: AddInlinedBodyRequest{
				F01: true,
				F02: 2,
				F03: 3,
				F04: 4,
				F05: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				F06: uuid.MustParse("b043c679-354a-4170-a061-dfe2271b3c77"),
				F07: "f7Test",
				F08: AddInlinedBodyRequestF08ValueB,
				F09: SeasonSummer,
				F10: []string{"a", "b"},
				F11: []int32{4, 5},
				F12: []Season{SeasonFall, SeasonSpring},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var actual AddInlinedBodyRequest
			err := json.Unmarshal([]byte(tc.Raw), &actual)
			if tc.ErrorMsg != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.ErrorMsg)
				return
			}

			assert.NoError(t, err)

			assert.Equal(t, tc.Expected, actual)
		})
	}
}
