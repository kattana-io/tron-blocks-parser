package tronApi

import (
	"go.uber.org/zap"
	"testing"
)

func TestApi_GetTokenDecimals(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	type fields struct {
		endpoint string
		log      *zap.Logger
		provider ApiUrlProvider
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "Basic case",
			fields: fields{
				endpoint: "https://api.trongrid.io",
				log:      logger,
				provider: NewTrongridUrlProvider(),
			},
			args: args{
				token: "4118fd0626daf3af02389aef3ed87db9c33f638ffa",
			},
			want:    18,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Api{
				endpoint: tt.fields.endpoint,
				log:      tt.fields.log,
				provider: tt.fields.provider,
			}
			got, err := a.GetTokenDecimals(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTokenDecimals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTokenDecimals() got = %v, want %v", got, tt.want)
			}
		})
	}
}
