package userCenter

import (
	"context"
	"fmt"
	"strconv"

	"github.com/a11en4sec/crawler/proto/user"
)

type LoginService struct{}

func (s *LoginService) UserLogin(ctx context.Context, req *user.UserLoginReq, resp *user.UserLoginResp) error {

	fmt.Printf("req name:%s, pwd:%s\n", req.Name, req.Pwd)

	if req.Name == "admin" && req.Pwd == "pwd" {

		s, _ := strconv.ParseInt(req.Sign, 10, 64)
		resp.Code = "1"
		resp.Msg = "login success"
		resp.Data = "ok"

		aToken, _, err := GenToken(s, req.Name)
		resp.Token = aToken

		if err != nil {
			return err
		}

	}

	return nil
}
