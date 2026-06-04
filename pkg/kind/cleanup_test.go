package kind

import (
	"errors"
	"testing"
)

func TestIsMissingContainerNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "docker named network not found",
			err:  errors.New("Error response from daemon: network dpu-sim-gateway not found"),
			want: true,
		},
		{
			name: "docker generic network not found",
			err:  errors.New("Error response from daemon: network not found"),
			want: true,
		},
		{
			name: "podman no such network",
			err:  errors.New("Error: no such network: dpu-sim-gateway"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("permission denied"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMissingContainerNetworkError(tt.err); got != tt.want {
				t.Fatalf("isMissingContainerNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}
