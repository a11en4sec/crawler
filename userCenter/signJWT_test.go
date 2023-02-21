package userCenter

import (
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestVerifyToken(t *testing.T) {
	type args struct {
		tokenID string
	}
	tests := []struct {
		name     string
		args     args
		want     *MyClaim
		userid   int64
		username string
		wantErr  bool
	}{
		{
			name: "test01",
			args: args{
				tokenID: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowLCJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNjc2NzE0MjMwfQ.1cMf7r4Y7AWRghiIjJj8jii7YpaYE2tIaukB4_q3cqg",
			},
			want: &MyClaim{
				UserID:           00000,
				Username:         "admin",
				RegisteredClaims: jwt.RegisteredClaims{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyToken(tt.args.tokenID)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
