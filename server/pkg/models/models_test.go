// server/pkg/models/models_test.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "string json",
			value: `{"key": "value"}`,
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "byte slice json",
			value: []byte(`{"key": "value"}`),
			want:  map[string]string{"key": "value"},
		},
		{
			name:    "nil value",
			value:   nil,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "invalid type",
			value:   123,
			wantErr: true,
		},
		{
			name:    "invalid json",
			value:   `{"invalid": json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSON[map[string]string]
			err := j.Scan(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.value != nil {
					assert.Equal(t, tt.want, j.Data)
				}
			}
		})
	}
}

func TestJSON_Value(t *testing.T) {
	j := JSON[map[string]string]{
		Data: map[string]string{"key": "value"},
	}

	value, err := j.Value()
	require.NoError(t, err)

	// Should return JSON bytes
	expected, _ := json.Marshal(j.Data)
	assert.Equal(t, expected, value)
}

func TestJSON_MarshalJSON(t *testing.T) {
	j := JSON[map[string]string]{
		Data: map[string]string{"key": "value"},
	}

	data, err := j.MarshalJSON()
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, j.Data, result)
}

func TestJSON_UnmarshalJSON(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	var j JSON[map[string]string]

	err := j.UnmarshalJSON(data)
	require.NoError(t, err)

	expected := map[string]string{"key": "value"}
	assert.Equal(t, expected, j.Data)
}

func TestJSON_DriverValueInterface(t *testing.T) {
	j := JSON[map[string]string]{
		Data: map[string]string{"key": "value"},
	}

	// Test that it implements driver.Valuer
	var _ driver.Valuer = j

	// Test Value() method
	value, err := j.Value()
	require.NoError(t, err)
	assert.NotNil(t, value)
}

func TestConstants(t *testing.T) {
	// Test states
	assert.Equal(t, "setup", Pending)
	assert.Equal(t, "building", Building)
	assert.Equal(t, "deploying", Deploying)
	assert.Equal(t, "success", Success)
	assert.Equal(t, "failed", Failed)

	// Test phases
	assert.Equal(t, "setup", Setup)
	assert.Equal(t, "build", Build)
	assert.Equal(t, "deploy", Deploy)
	assert.Equal(t, "ssl", SSL)
	assert.Equal(t, "complete", Complete)
}
