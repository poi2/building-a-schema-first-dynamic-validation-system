package version

import "testing"

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int32
		wantMinor int32
		wantPatch int32
		wantErr   bool
	}{
		{
			name:      "valid version",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "zero version",
			version:   "0.0.0",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:    "prerelease version",
			version: "1.2.3-RC-1",
			wantErr: true,
		},
		{
			name:    "version with v prefix",
			version: "v1.2.3",
			wantErr: true,
		},
		{
			name:    "too few parts",
			version: "1.2",
			wantErr: true,
		},
		{
			name:    "too many parts",
			version: "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "non-numeric major",
			version: "a.2.3",
			wantErr: true,
		},
		{
			name:    "non-numeric minor",
			version: "1.b.3",
			wantErr: true,
		},
		{
			name:    "non-numeric patch",
			version: "1.2.c",
			wantErr: true,
		},
		{
			name:    "negative version",
			version: "1.-2.3",
			wantErr: true,
		},
		{
			name:    "int32 overflow major",
			version: "2147483648.0.0",
			wantErr: true,
		},
		{
			name:    "int32 overflow minor",
			version: "0.2147483648.0",
			wantErr: true,
		},
		{
			name:    "int32 overflow patch",
			version: "0.0.2147483648",
			wantErr: true,
		},
		{
			name:      "int32 max valid",
			version:   "2147483647.2147483647.2147483647",
			wantMajor: 2147483647,
			wantMinor: 2147483647,
			wantPatch: 2147483647,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := ParseSemVer(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSemVer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor {
					t.Errorf("ParseSemVer() major = %v, want %v", major, tt.wantMajor)
				}
				if minor != tt.wantMinor {
					t.Errorf("ParseSemVer() minor = %v, want %v", minor, tt.wantMinor)
				}
				if patch != tt.wantPatch {
					t.Errorf("ParseSemVer() patch = %v, want %v", patch, tt.wantPatch)
				}
			}
		})
	}
}
