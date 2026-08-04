package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestVirtualPaymentNotifyParsesPlainMessage(t *testing.T) {
	config := viper.New()
	config.Set("virtual_payment.message_token", "notify-token")
	config.Set("virtual_payment.env", 1)
	service := NewVirtualPaymentNotifyService(config)

	timestamp, nonce := "1720000000", "nonce"
	signature := testNotifySignature("notify-token", timestamp, nonce)
	body := []byte(`<xml><MsgType>event</MsgType><Event>xpay_goods_deliver_notify</Event><OpenId>openid</OpenId><OutTradeNo>RENT202608010001</OutTradeNo><Env>1</Env><WeChatPayInfo><TransactionId>wxtx001</TransactionId></WeChatPayInfo><GoodsInfo><ProductId>rent_publish_1</ProductId><Quantity>1</Quantity><ActualPrice>1800</ActualPrice></GoodsInfo></xml>`)
	notice, err := service.Parse(signature, timestamp, nonce, body)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if notice.OutTradeNo != "RENT202608010001" || notice.TransactionID != "wxtx001" || notice.Environment != 1 || notice.ProductID != "rent_publish_1" || notice.ActualPriceCents != 1800 {
		t.Fatalf("unexpected notice: %#v", notice)
	}
}

func TestVirtualPaymentNotifyRejectsWrongEnvironment(t *testing.T) {
	config := viper.New()
	config.Set("virtual_payment.message_token", "notify-token")
	config.Set("virtual_payment.env", 0)
	service := NewVirtualPaymentNotifyService(config)
	body := []byte(`<xml><MsgType>event</MsgType><Event>xpay_goods_deliver_notify</Event><OpenId>openid</OpenId><OutTradeNo>RENT1</OutTradeNo><Env>1</Env><WeChatPayInfo><TransactionId>wxtx001</TransactionId></WeChatPayInfo><GoodsInfo><ProductId>rent_publish_1</ProductId><Quantity>1</Quantity><ActualPrice>1800</ActualPrice></GoodsInfo></xml>`)
	_, err := service.Parse(testNotifySignature("notify-token", "1", "2"), "1", "2", body)
	if err != ErrVirtualPaymentNotifyPayload {
		t.Fatalf("Parse() error = %v, want %v", err, ErrVirtualPaymentNotifyPayload)
	}
}

func TestVirtualPaymentNotifyDecryptsSafeMode(t *testing.T) {
	const (
		token = "notify-token"
		appID = "wx-test"
		key   = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"
	)
	config := viper.New()
	config.Set("virtual_payment.message_token", token)
	config.Set("virtual_payment.encoding_aes_key", key)
	config.Set("virtual_payment.env", 0)
	config.Set("wechat.app_id", appID)
	service := NewVirtualPaymentNotifyService(config)

	plain := `<xml><MsgType>event</MsgType><Event>xpay_goods_deliver_notify</Event><OpenId>openid</OpenId><OutTradeNo>RENT1</OutTradeNo><Env>0</Env><WeChatPayInfo><TransactionId>wxtx001</TransactionId></WeChatPayInfo><GoodsInfo><ProductId>rent_publish_1</ProductId><Quantity>1</Quantity><ActualPrice>1800</ActualPrice></GoodsInfo></xml>`
	encrypted := encryptTestWechatMessage(t, key, appID, plain)
	timestamp, nonce := "3", "4"
	signature := testNotifySignature(token, timestamp, nonce, encrypted)
	body := []byte("<xml><Encrypt><![CDATA[" + encrypted + "]]></Encrypt></xml>")
	notice, err := service.Parse(signature, timestamp, nonce, body)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if notice.TransactionID != "wxtx001" {
		t.Fatalf("unexpected notice: %#v", notice)
	}
}

func testNotifySignature(token string, values ...string) string {
	parts := append([]string{token}, values...)
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(sum[:])
}

func encryptTestWechatMessage(t *testing.T, encodingAESKey, appID, message string) string {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		t.Fatal(err)
	}
	payload := make([]byte, 20+len(message)+len(appID))
	copy(payload[:16], []byte("0123456789abcdef"))
	binary.BigEndian.PutUint32(payload[16:20], uint32(len(message)))
	copy(payload[20:], message)
	copy(payload[20+len(message):], appID)
	padding := aes.BlockSize - len(payload)%aes.BlockSize
	payload = append(payload, bytesRepeat(byte(padding), padding)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := make([]byte, len(payload))
	cipher.NewCBCEncrypter(block, key[:aes.BlockSize]).CryptBlocks(ciphertext, payload)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func bytesRepeat(value byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = value
	}
	return result
}
