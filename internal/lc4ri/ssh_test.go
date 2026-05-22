package lc4ri

import "testing"

func TestValidatePublicKeyLine(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name: "ed25519",
			key:  "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICZ comment",
		},
		{
			name: "security key",
			key:  "sk-ssh-ed25519@openssh.com AAAAC3NzaC1lZDI1NTE5AAAAICZ comment",
		},
		{
			name:    "authorized keys option",
			key:     `command="echo hacked" ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICZ`,
			wantErr: true,
		},
		{
			name:    "missing key data",
			key:     "ssh-ed25519",
			wantErr: true,
		},
		{
			name:    "unknown algorithm",
			key:     "ssh-dss AAAAB3NzaC1kc3MAAACB",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePublicKeyLine(tt.key)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
