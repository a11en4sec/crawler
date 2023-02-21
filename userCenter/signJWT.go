package userCenter

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"go-micro.dev/v4/auth"
)

const (
	ATokenExpiredDuration = 2 * time.Hour
	RTokenExpiredDuration = 30 * 24 * time.Hour
	TokenIssuer           = ""
)

var (
	mySecret          = []byte("xxxx")
	ErrorInvalidToken = errors.New("verify Token Failed")
)

type MyClaim struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
	opts auth.Options
}

func (m *MyClaim) Init(opts ...auth.Option) {
	options := auth.Options{}
	for _, opt := range opts {
		opt(&options)
	}
}

func (m *MyClaim) Options() auth.Options {
	return m.opts
}

func (m *MyClaim) Generate(id string, opts ...auth.GenerateOption) (*auth.Account, error) {

	options := auth.NewGenerateOptions(opts...)

	return &auth.Account{
		ID:       id,
		Secret:   options.Secret,
		Metadata: options.Metadata,
		Scopes:   options.Scopes,
		Issuer:   m.Options().Namespace,
	}, nil
}

func (m *MyClaim) Inspect(tokenStr string) (*auth.Account, error) {
	token, err := jwt.ParseWithClaims(tokenStr, m, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		err = ErrorInvalidToken
		return nil, err
	}
	return m.Generate(m.ID)
}

func (m *MyClaim) Token(opts ...auth.TokenOption) (*auth.Token, error) {

	rc := jwt.RegisteredClaims{
		ExpiresAt: getJWTTime(ATokenExpiredDuration),
		Issuer:    TokenIssuer,
	}

	atoken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, m).SignedString(mySecret)

	// refresh token 不需要保存任何用户信息
	rt := rc
	rt.ExpiresAt = getJWTTime(RTokenExpiredDuration)
	rtoken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rt).SignedString(mySecret)

	return &auth.Token{
		AccessToken:  atoken,
		RefreshToken: rtoken,
		Created:      time.Now(),
		Expiry:       rc.ExpiresAt.Time,
	}, nil

}

func (m *MyClaim) String() string {
	//TODO implement me
	//panic("implement me")
	return "userCenter"
}

func getJWTTime(t time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(t))
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	return mySecret, nil
}

// GenToken 颁发token: access token 和 refresh token
func GenToken(UserID int64, Username string) (atoken, rtoken string, err error) {
	rc := jwt.RegisteredClaims{
		ExpiresAt: getJWTTime(ATokenExpiredDuration),
		Issuer:    TokenIssuer,
	}
	at := MyClaim{
		UserID:           UserID,
		Username:         Username,
		RegisteredClaims: rc,
	}
	atoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, at).SignedString(mySecret)

	// refresh token 不需要保存任何用户信息
	rt := rc
	rt.ExpiresAt = getJWTTime(RTokenExpiredDuration)
	rtoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, rt).SignedString(mySecret)
	return
}

// VerifyToken 验证Token
func VerifyToken(tokenID string) (*MyClaim, error) {
	var myc = new(MyClaim)
	token, err := jwt.ParseWithClaims(tokenID, myc, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		err = ErrorInvalidToken
		return nil, err
	}

	return myc, nil
}

// RefreshToken 通过 refresh token 刷新 atoken
func RefreshToken(atoken, rtoken string) (newAtoken, newRtoken string, err error) {
	// rtoken 无效直接返回
	if _, err = jwt.Parse(rtoken, keyFunc); err != nil {
		return
	}
	// 从旧access token 中解析出claims数据
	var claim MyClaim
	_, err = jwt.ParseWithClaims(atoken, &claim, keyFunc)
	// 判断错误是不是因为access token 正常过期导致的
	v, _ := err.(*jwt.ValidationError)
	if v.Errors == jwt.ValidationErrorExpired {
		return GenToken(claim.UserID, claim.Username)
	}
	return
}
