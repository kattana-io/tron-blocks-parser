package parser

import "testing"

func Test_getMethodId(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Transfer event",
			args: args{
				input: "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
			},
			want: 0xddf252ad,
		},
		{
			name: "Token purchase event",
			args: args{
				input: "cd60aa75dea3072fbc07ae6d7d856b5dc5f4eee88854f5b4abf7b680ef8bc50f",
			},
			want: 0xcd60aa75,
		},
		{
			name: "Snapshot event",
			args: args{
				input: "cc7244d3535e7639366f8c5211527112e01de3ec7449ee3a6e66b007f4065a70",
			},
			want: 0xcc7244d3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMethodId(tt.args.input); got != tt.want {
				t.Errorf("getMethodId() = %v, want %v", got, tt.want)
			}
		})
	}
}
