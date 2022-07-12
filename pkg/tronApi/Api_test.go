package tronApi

import "testing"

func Test_normalizeAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Basic case",
			args: args{
				address: "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
			},
			want: "41a2726afbecbd8e936000ed684cef5e2f5cf43008",
		},
		{
			name: "Invalid case",
			args: args{
				address: "0x123",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAddress(tt.args.address); got != tt.want {
				t.Errorf("normalizeAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimZeroes(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Basic case",
			args: args{
				address: "000000000000000000000000a614f803b6fd780986a42c78ec9c7f77e6ded13c",
			},
			want: "a614f803b6fd780986a42c78ec9c7f77e6ded13c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimZeroes(tt.args.address); got != tt.want {
				t.Errorf("trimZeroes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeAddress(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Basic case",
			args: args{
				raw: "41a2726afbecbd8e936000ed684cef5e2f5cf43008",
			},
			want: "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeAddress(tt.args.raw); got != tt.want {
				t.Errorf("decodeAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
