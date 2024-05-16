package client

import (
	"testing"
)

func TestGetResourceState(t *testing.T) {
	tests := []struct {
		name           string
		resourceState  int
		clientState    int
		requestedState int
		want           bool
	}{
		{
			name:           "resourceState is -1, clientState is -1, requestedState is 1",
			resourceState:  -1,
			clientState:    -1,
			requestedState: StateCreateIfNecessary,
			want:           false,
		},
		{
			name:           "resourceState is 1, clientState is -1, requestedState is 1",
			resourceState:  1,
			clientState:    -1,
			requestedState: StateCreateIfNecessary,
			want:           true,
		},
		{
			name:           "resourceState is -1, clientState is 1, requestedState is 1",
			resourceState:  -1,
			clientState:    1,
			requestedState: StateCreateIfNecessary,
			want:           true,
		},
		{
			name:           "resourceState is 1, clientState is 0, requestedState is 1",
			resourceState:  1,
			clientState:    1,
			requestedState: StateCreateIfNecessary,
			want:           true,
		},
		{
			name:           "resourceState is 3, clientState is 0, requestedState is 2",
			resourceState:  3,
			clientState:    0,
			requestedState: StateAlwaysDelete,
			want:           true,
		},
		{
			name:           "resourceState is 1, clientState is 1, requestedState is 2",
			resourceState:  1,
			clientState:    1,
			requestedState: StateAlwaysDelete,
			want:           false,
		},
		{
			name:           "resourceState is -1, clientState is 3, requestedState is 2",
			resourceState:  -1,
			clientState:    3,
			requestedState: StateAlwaysDelete,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getResourceState(tt.resourceState, tt.clientState, tt.requestedState); got != tt.want {
				t.Errorf("getResourceState() = %v, want %v", got, tt.want)
			}
		})
	}
}
