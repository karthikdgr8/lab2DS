package control

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
)

var Key []byte
var err error

func SHAify(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	//base64String := base64.URLEncoding.EncodeToString(hasher.Sum(nil))[0:8]
	base64Int := new(big.Int).SetBytes(hasher.Sum(nil)[0:8])
	base16String := base64Int.Text(16)
	return base16String
}

func getKeyFromFile() {
	// The Key should be 16 bytes (AES-128), 24 bytes (AES-192) or
	// 32 bytes (AES-256)
	Key, err = os.ReadFile("/Users/karthik/Downloads/key.txt")
	if err != nil {
		log.Fatal(err)
	}
}

func fileEncryptAndSend(filePath string) []byte {

	file, err := os.ReadFile(filePath)

	block, err := aes.NewCipher(Key)
	if err != nil {
		log.Panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic(err)
	}

	// Never use more than 2^32 random nonces with a given Key
	// because of the risk of repeat.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, file, nil)

	return ciphertext
}

func fileDecryptAndSend(fileName string) []byte {

	ciphertext, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	block, err := aes.NewCipher(Key)
	if err != nil {
		log.Panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic(err)
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Panic(err)
	}

	return plaintext

}

func encryptAESConn(sendData []byte) ([]byte, []byte) {

	block, cipherErr := aes.NewCipher(Key)

	if cipherErr != nil {
		fmt.Errorf("Can't create cipher:", cipherErr)

		return nil, nil
	}

	iv := make([]byte, aes.BlockSize)

	if _, randReadErr := io.ReadFull(rand.Reader, iv); randReadErr != nil {
		fmt.Errorf("Can't build random iv", randReadErr)
		return nil, nil
	}

	stream := cipher.NewCFBEncrypter(block, iv)

	encrypted := make([]byte, len(sendData))

	stream.XORKeyStream(encrypted, sendData)

	return iv, encrypted

}

func decryptAESConn(iv []byte, recvData []byte) []byte {

	block, cipherErr := aes.NewCipher(Key)

	if cipherErr != nil {
		fmt.Errorf("Can't create cipher:", cipherErr)

		return nil
	}

	stream := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, len(recvData))

	stream.XORKeyStream(decrypted, recvData)

	return decrypted

}
