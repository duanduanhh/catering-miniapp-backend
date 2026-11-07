package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

const virtualPaymentGoodsDeliverEvent = "xpay_goods_deliver_notify"

var (
	ErrVirtualPaymentNotifyUnavailable = errors.New("virtual payment message push is not configured")
	ErrVirtualPaymentNotifySignature   = errors.New("invalid virtual payment notification signature")
	ErrVirtualPaymentNotifyPayload     = errors.New("invalid virtual payment notification payload")
)

// VirtualPaymentNotifyService 负责验证、解密微信虚拟支付消息推送。
// 业务发货由 handler 在通知验证通过后调用订单服务完成。
type VirtualPaymentNotifyService interface {
	VerifyURL(signature, timestamp, nonce, echo string) (string, error)
	Parse(signature, timestamp, nonce string, body []byte) (*VirtualPaymentGoodsDelivery, error)
}

type VirtualPaymentGoodsDelivery struct {
	OutTradeNo       string
	OpenID           string
	Environment      int
	TransactionID    string
	ProductID        string
	Quantity         int
	ActualPriceCents int64
}

func NewVirtualPaymentNotifyService(config *viper.Viper) VirtualPaymentNotifyService {
	return &virtualPaymentNotifyService{
		token:          strings.TrimSpace(config.GetString("virtual_payment.message_token")),
		encodingAESKey: strings.TrimSpace(config.GetString("virtual_payment.encoding_aes_key")),
		appID:          strings.TrimSpace(config.GetString("wechat.app_id")),
		environment:    config.GetInt("virtual_payment.env"),
	}
}

type virtualPaymentNotifyService struct {
	token          string
	encodingAESKey string
	appID          string
	environment    int
}

func (s *virtualPaymentNotifyService) VerifyURL(signature, timestamp, nonce, echo string) (string, error) {
	if s.token == "" {
		return "", ErrVirtualPaymentNotifyUnavailable
	}
	// In plaintext mode WeChat signs token, timestamp and nonce only. The
	// console still requires an EncodingAESKey to be configured, so its presence
	// alone must not make this validation take the encrypted-message path.
	if validSignature(s.token, signature, timestamp, nonce) {
		return echo, nil
	}

	// In safe mode echostr is encrypted and participates in the message
	// signature. Only decrypt it after that signature has been verified.
	if s.encodingAESKey != "" && validSignature(s.token, signature, timestamp, nonce, echo) {
		return s.decrypt(echo)
	}
	return "", ErrVirtualPaymentNotifySignature
}

func (s *virtualPaymentNotifyService) Parse(signature, timestamp, nonce string, body []byte) (*VirtualPaymentGoodsDelivery, error) {
	if s.token == "" {
		return nil, ErrVirtualPaymentNotifyUnavailable
	}
	var envelope wechatMessageEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVirtualPaymentNotifyPayload, err)
	}
	messageXML := body
	if envelope.Encrypt != "" {
		if !validSignature(s.token, signature, timestamp, nonce, envelope.Encrypt) {
			return nil, ErrVirtualPaymentNotifySignature
		}
		decrypted, err := s.decrypt(envelope.Encrypt)
		if err != nil {
			return nil, err
		}
		messageXML = []byte(decrypted)
	} else if !validSignature(s.token, signature, timestamp, nonce) {
		return nil, ErrVirtualPaymentNotifySignature
	}

	var message virtualPaymentMessage
	if err := xml.Unmarshal(messageXML, &message); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVirtualPaymentNotifyPayload, err)
	}
	if message.MsgType != "event" || message.Event != virtualPaymentGoodsDeliverEvent || message.OutTradeNo == "" ||
		message.OpenID == "" || message.GoodsInfo.ProductID == "" || message.GoodsInfo.Quantity <= 0 || message.GoodsInfo.ActualPrice < 0 {
		return nil, ErrVirtualPaymentNotifyPayload
	}
	if message.Env != s.environment {
		return nil, ErrVirtualPaymentNotifyPayload
	}
	transactionID := message.WeChatPayInfo.TransactionID
	if transactionID == "" {
		transactionID = message.WeChatPayInfo.MchOrderNo
	}
	if transactionID == "" {
		return nil, ErrVirtualPaymentNotifyPayload
	}
	return &VirtualPaymentGoodsDelivery{
		OutTradeNo:       message.OutTradeNo,
		OpenID:           message.OpenID,
		Environment:      message.Env,
		TransactionID:    transactionID,
		ProductID:        message.GoodsInfo.ProductID,
		Quantity:         message.GoodsInfo.Quantity,
		ActualPriceCents: message.GoodsInfo.ActualPrice,
	}, nil
}

func (s *virtualPaymentNotifyService) decrypt(encrypted string) (string, error) {
	if s.encodingAESKey == "" || s.appID == "" {
		return "", ErrVirtualPaymentNotifyUnavailable
	}
	key, err := base64.StdEncoding.DecodeString(s.encodingAESKey + "=")
	if err != nil || len(key) != 32 {
		return "", ErrVirtualPaymentNotifyUnavailable
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil || len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", ErrVirtualPaymentNotifyPayload
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, key[:aes.BlockSize]).CryptBlocks(plain, ciphertext)
	plain, err = unpadPKCS7(plain)
	if err != nil || len(plain) < 20 {
		return "", ErrVirtualPaymentNotifyPayload
	}
	messageLength := int(binary.BigEndian.Uint32(plain[16:20]))
	if messageLength < 0 || 20+messageLength > len(plain) {
		return "", ErrVirtualPaymentNotifyPayload
	}
	message := string(plain[20 : 20+messageLength])
	if appID := string(plain[20+messageLength:]); appID != s.appID {
		return "", ErrVirtualPaymentNotifyPayload
	}
	return message, nil
}

func validSignature(token, signature string, values ...string) bool {
	parts := append([]string{token}, values...)
	sort.Strings(parts)
	digest := sha1.Sum([]byte(strings.Join(parts, "")))
	return strings.EqualFold(signature, hex.EncodeToString(digest[:]))
}

func unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty padded data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) || !bytes.Equal(data[len(data)-padding:], bytes.Repeat([]byte{byte(padding)}, padding)) {
		return nil, errors.New("invalid PKCS7 padding")
	}
	return data[:len(data)-padding], nil
}

type wechatMessageEnvelope struct {
	Encrypt string `xml:"Encrypt"`
}

type virtualPaymentMessage struct {
	MsgType       string `xml:"MsgType"`
	Event         string `xml:"Event"`
	OpenID        string `xml:"OpenId"`
	OutTradeNo    string `xml:"OutTradeNo"`
	Env           int    `xml:"Env"`
	WeChatPayInfo struct {
		MchOrderNo    string `xml:"MchOrderNo"`
		TransactionID string `xml:"TransactionId"`
	} `xml:"WeChatPayInfo"`
	GoodsInfo struct {
		ProductID   string `xml:"ProductId"`
		Quantity    int    `xml:"Quantity"`
		ActualPrice int64  `xml:"ActualPrice"`
	} `xml:"GoodsInfo"`
}
