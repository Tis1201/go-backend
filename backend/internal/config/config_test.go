package config

import "testing"

func fakeEnv(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		wantErr  bool
		wantAddr string
	}{
		{name: "missing DATABASE_URL", env: map[string]string{}, wantErr: true},
		{name: "empty DATABASE_URL", env: map[string]string{"DATABASE_URL": ""}, wantErr: true},
		{name: "default addr", env: map[string]string{"DATABASE_URL": "postgres://x"}, wantAddr: ":8080"},
		{name: "custom addr", env: map[string]string{"DATABASE_URL": "postgres://x", "HTTP_ADDR": ":9090"}, wantAddr: ":9090"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(fakeEnv(tc.env))
			if tc.wantErr {
				if err == nil {
					t.Fatal("Load() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if cfg.HTTPAddr != tc.wantAddr {
				t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, tc.wantAddr)
			}
		})
	}
}
