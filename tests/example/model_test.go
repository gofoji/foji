package example

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func ToPointer[T any](v T) *T {
	return &v
}

func TestAddInlinedBodyRequest(t *testing.T) {
	testCases := []struct {
		Name     string
		Raw      string
		Expected AddInlinedBodyRequest
		ErrorMsg string
	}{
		{
			Name: "minimal",
			Raw:  `{"f07": "test", "f10": ["a"]}`,
			Expected: AddInlinedBodyRequest{
				F01B:     true,
				F01BNull: ToPointer(true),
				F04:      1,
				F04Null:  ToPointer(int64(2)),
				F07:      "test",
				F08:      AddInlinedBodyRequestF08ValueA,
				F08Null:  ToPointer(AddInlinedBodyRequestF08NullValueB),
				F10:      []string{"a"},
				F13:      "someValue",
				F13Null:  ToPointer("someValue2"),
			},
		},
		{
			Name: "maximal",
			Raw: `{"f01": true, "f01Null": true, "f02": 2, "f02Null": 21, "f03": 3, "f03Null": 31, "f04": 4, "f04Null": 41, 
"f05": "2025-01-01T12:00:00Z", "f05Null": "2025-02-01T12:00:00Z", 
"f06": "b043c679-354a-4170-a061-dfe2271b3c77", "f06Null": "f4e1d34e-231a-4e4d-a904-318e8edd5a29", 
"f07": "f7Test", "f07Null": "f7NullTest", "f08": "valueB", "f08Null": "valueC", "f09": "summer", "f09Null": "fall", 
"f10": ["a", "b"], "f11": [4, 5], "f12": ["fall", "spring"], "f13": "f13Test", "f13Null": "f13NullTest"}`,
			Expected: AddInlinedBodyRequest{
				F01:      true,
				F01Null:  ToPointer(true),
				F01B:     true,
				F01BNull: ToPointer(true),
				F02:      2,
				F02Null:  ToPointer(int32(21)),
				F03:      3,
				F03Null:  ToPointer(int32(31)),
				F04:      4,
				F04Null:  ToPointer(int64(41)),
				F05:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				F05Null:  ToPointer(time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)),
				F06:      uuid.MustParse("b043c679-354a-4170-a061-dfe2271b3c77"),
				F06Null:  ToPointer(uuid.MustParse("f4e1d34e-231a-4e4d-a904-318e8edd5a29")),
				F07:      "f7Test",
				F07Null:  ToPointer("f7NullTest"),
				F08:      AddInlinedBodyRequestF08ValueB,
				F08Null:  ToPointer(AddInlinedBodyRequestF08NullValueC),
				F09:      SeasonSummer,
				F09Null:  ToPointer(SeasonNullableFall),
				F10:      []string{"a", "b"},
				F11:      []int32{4, 5},
				F12:      []Season{SeasonFall, SeasonSpring},
				F13:      "f13Test",
				F13Null:  ToPointer("f13NullTest"),
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

func TestAddFormRequest(t *testing.T) {
	testCases := []struct {
		Name     string
		Data     url.Values
		Expected AddFormRequest
		ErrorMsg string
	}{
		{
			Name: "minimal",
			Data: url.Values{
				"f07": {"f7Test"},
				"f10": {"a,b"},
			},
			Expected: AddFormRequest{
				F01B:     true,
				F01BNull: ToPointer(true),
				F04:      1,
				F04Null:  ToPointer(int64(2)),
				F07:      "f7Test",
				F08:      AddFormRequestF08ValueA,
				F08Null:  ToPointer(AddFormRequestF08NullValueB),
				F10:      []string{"a", "b"},
				F13:      "someValue",
				F13Null:  ToPointer("someValue2"),
			},
		},
		{
			Name: "maximal",
			Data: url.Values{
				"f01":      {"true"},
				"f01Null":  {"true"},
				"f01b":     {"true"},
				"f01bNull": {"true"},
				"f02":      {"2"},
				"f02Null":  {"21"},
				"f03":      {"3"},
				"f03Null":  {"31"},
				"f04":      {"4"},
				"f04Null":  {"41"},
				"f05":      {"2025-01-01T12:00:00Z"},
				"f05Null":  {"2025-02-01T12:00:00Z"},
				"f06":      {"b043c679-354a-4170-a061-dfe2271b3c77"},
				"f06Null":  {"f4e1d34e-231a-4e4d-a904-318e8edd5a29"},
				"f07":      {"f7Test"},
				"f07Null":  {"f7NullTest"},
				"f08":      {"valueB"},
				"f08Null":  {"valueC"},
				"f09":      {"summer"},
				"f09Null":  {"fall"},
				"f10":      {"a,b"},
				"f11":      {"4,5"},
				"f12":      {"fall,spring"},
				"f13":      {"f13Test"},
				"f13Null":  {"f13NullTest"},
			},
			Expected: AddFormRequest{
				F01:      true,
				F01Null:  ToPointer(true),
				F01B:     true,
				F01BNull: ToPointer(true),
				F02:      2,
				F02Null:  ToPointer(int32(21)),
				F03:      3,
				F03Null:  ToPointer(int32(31)),
				F04:      4,
				F04Null:  ToPointer(int64(41)),
				F05:      time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				F05Null:  ToPointer(time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)),
				F06:      uuid.MustParse("b043c679-354a-4170-a061-dfe2271b3c77"),
				F06Null:  ToPointer(uuid.MustParse("f4e1d34e-231a-4e4d-a904-318e8edd5a29")),
				F07:      "f7Test",
				F07Null:  ToPointer("f7NullTest"),
				F08:      AddFormRequestF08ValueB,
				F08Null:  ToPointer(AddFormRequestF08NullValueC),
				F09:      SeasonSummer,
				F09Null:  ToPointer(SeasonNullableFall),
				F10:      []string{"a", "b"},
				F11:      []int32{4, 5},
				F12:      []Season{SeasonFall, SeasonSpring},
				F13:      "f13Test",
				F13Null:  ToPointer("f13NullTest"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "test", strings.NewReader(tc.Data.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			actual, err := ParseFormAddFormRequest(req)
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
