package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/spf13/viper"
)

type PayService interface {
	BuildJSAPIPayParams(ctx context.Context, orderNo string, amount float64) (v1.PayParams, error)
}

type payService struct {
	config *viper.Viper
}

func NewPayService(config *viper.Viper) PayService {
	return &payService{config: config}
}

func (s *payService) BuildJSAPIPayParams(ctx context.Context, orderNo string, amount float64) (v1.PayParams, error) {
	appID := s.config.GetString("wxpay.app_id")
	if appID == "" {
		return v1.PayParams{}, errors.New("wxpay.app_id is empty")
	}
	nonce := strconv.FormatInt(time.Now().UnixNano(), 36)
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	pkg := fmt.Sprintf("prepay_id=%s", orderNo)
	signType := "RSA"
	paySign := signPayParams(appID, timeStamp, nonce, pkg)
	return v1.PayParams{
		TimeStamp: timeStamp,
		NonceStr:  nonce,
		Package:   pkg,
		SignType:  signType,
		PaySign:   paySign,
	}, nil
}

func signPayParams(appID, timeStamp, nonceStr, pkg string) string {
	payload := strings.Join([]string{appID, timeStamp, nonceStr, pkg}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
