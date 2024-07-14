package service

import "testing"

func TestValidateLuhn(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Positive #1",
			args:    args{number: "3564005076"},
			wantErr: false,
		},
		{
			name:    "Positive #2",
			args:    args{number: "1708232002"},
			wantErr: false,
		},
		{
			name:    "Negative #1",
			args:    args{number: "13564005076"},
			wantErr: true,
		},
		{
			name:    "Negative #2",
			args:    args{number: "14xyz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateLuhn(tt.args.number); (err != nil) != tt.wantErr {
				t.Errorf("ValidateLuhn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
