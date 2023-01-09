package sec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
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

func GetKeyFromFile() {
	// The Key should be 16 bytes (AES-128), 24 bytes (AES-192) or
	// 32 bytes (AES-256)
	Key, err = os.ReadFile("./files/key.txt")
	if err != nil {
		log.Fatal(err)
	}
}

func GeneratePrivate() *ecdsa.PrivateKey {
	ret, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return ret
}

func CalculateSessionKey(priv ecdsa.PrivateKey, X, Y *big.Int) []byte {

	//priva, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	//privb, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	//puba := priva.PublicKey
	//pubb := privb.PublicKey

	//fmt.Printf("\nPrivate key (Alice) %x", priva.D)
	//fmt.Printf("\nPrivate key (Bob) %x\n", privb.D)

	//fmt.Printf("\nPublic key (Alice) (%x,%x)", puba.X, puba.Y)
	//fmt.Printf("\nPublic key (Bob) (%x %x)\n", pubb.X, pubb.Y)
	pub := priv.PublicKey
	a, _ := pub.Curve.ScalarMult(X, Y, priv.D.Bytes())
	//b, _ := pubb.Curve.ScalarMult(pubb.X, pubb.Y, priva.D.Bytes())

	shared1 := sha256.Sum256(a.Bytes())
	//shared2 := sha256.Sum256(b.Bytes())

	fmt.Printf("\nShared key (Alice) %x\n", shared1)
	return shared1[:]

}

func FileEncryptAndSend(filePath string) []byte {

	file, err := os.ReadFile(filePath)

	if err != nil {
		return nil
	}

	return Encrypt(Key, file) // Change Key from reading file to Diffie Hellman
}

func Encrypt(key, data []byte) []byte {
	block, err := aes.NewCipher(key)
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

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext
}

func Decrypt(key, data []byte) []byte {

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic(err)
	}

	nonce := data[:gcm.NonceSize()]
	data = data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		log.Panic(err)
	}

	return plaintext
}

func FileDecryptAndSend(fileName string) []byte {

	ciphertext, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return Decrypt(Key, ciphertext)

}

func EncryptAESConn(sendData []byte) ([]byte, []byte) {

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
