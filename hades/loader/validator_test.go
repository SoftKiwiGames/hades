package loader

import (
	"testing"

	"github.com/SoftKiwiGames/hades/hades/schema"
)

func TestValidateEnvContract(t *testing.T) {
	tests := []struct {
		name     string
		job      schema.Job
		provided map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "all required vars provided",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
					"MODE":    {Default: "prod"},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
			},
			wantErr: false,
		},
		{
			name: "missing required var",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
					"MODE":    {},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
			},
			wantErr: true,
			errMsg:  "required environment variable \"MODE\" not provided",
		},
		{
			name: "unknown var provided is ignored",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
				"UNKNOWN": "value",
			},
			wantErr: false,
		},
		{
			name: "user provides HADES_ var",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
				},
			},
			provided: map[string]string{
				"VERSION":       "v1.0.0",
				"HADES_RUN_ID": "123",
			},
			wantErr: true,
			errMsg:  "user cannot provide HADES_* environment variables",
		},
		{
			name: "job defines HADES_ var",
			job: schema.Job{
				Env: map[string]schema.Env{
					"HADES_CUSTOM": {},
				},
			},
			provided: map[string]string{},
			wantErr:  true,
			errMsg:   "job cannot define HADES_* environment variables",
		},
		{
			name: "optional var not provided - ok",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
					"MODE":    {Default: "prod"},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvContract(&tt.job, tt.provided)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEnvContract() error = %v, want substring %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestMergeEnv(t *testing.T) {
	tests := []struct {
		name     string
		job      schema.Job
		provided map[string]string
		want     map[string]string
	}{
		{
			name: "provided overrides default",
			job: schema.Job{
				Env: map[string]schema.Env{
					"MODE": {Default: "prod"},
				},
			},
			provided: map[string]string{
				"MODE": "staging",
			},
			want: map[string]string{
				"MODE": "staging",
			},
		},
		{
			name: "default used when not provided",
			job: schema.Job{
				Env: map[string]schema.Env{
					"MODE":    {Default: "prod"},
					"VERSION": {},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
			},
			want: map[string]string{
				"MODE":    "prod",
				"VERSION": "v1.0.0",
			},
		},
		{
			name: "required var without default",
			job: schema.Job{
				Env: map[string]schema.Env{
					"VERSION": {},
				},
			},
			provided: map[string]string{
				"VERSION": "v1.0.0",
			},
			want: map[string]string{
				"VERSION": "v1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeEnv(&tt.job, tt.provided)
			if len(got) != len(tt.want) {
				t.Errorf("MergeEnv() length = %v, want %v", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("MergeEnv()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
