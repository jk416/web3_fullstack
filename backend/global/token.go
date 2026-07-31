package global

import (
	"time"
	"web3-wallet-exchange/token"
)

var Token token.TokenMaker // ← 类型是接口 TokenMaker，不是 *JWTMaker

func InitToken() {
	Token = token.NewJWTMaker(Conf.JWT.Secret, time.Duration(Conf.JWT.ExpireHours)*time.Hour)
}
