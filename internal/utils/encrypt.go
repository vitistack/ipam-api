package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/spf13/viper"
)

func DeterministicEncrypt(plaintext string) (string, error) {
	encryptionKey := viper.GetString("enc_key")
	encryptionIv := viper.GetString("enc_iv")
	var key = []byte(encryptionKey)
	var fixedNonce = []byte(encryptionIv)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, fixedNonce)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DeterministicDecrypt(encoded string) (string, error) {
	encryptionKey := viper.GetString("enc_key")
	encryptionIv := viper.GetString("enc_iv")
	var key = []byte(encryptionKey)
	var fixedNonce = []byte(encryptionIv)

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, fixedNonce)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
