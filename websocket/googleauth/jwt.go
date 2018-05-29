package googleauth

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"

	. "Redis-Exploration/websocket/models"

	uuid "github.com/satori/go.uuid"
)

type JWTAuthentication struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

const (
	tokenDuration = 72
	expireOffset  = 3600
)

var jwtAuthentication *JWTAuthentication = nil

func getPrivateKey() *rsa.PrivateKey {
	abs, err := filepath.Abs("./googleauth/private_key.txt")
	if err != nil {
		panic(err)
	}

	privateKeyFile, err := os.Open(abs)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	abs, err := filepath.Abs("./googleauth/public_key.pub")
	if err != nil {
		panic(err)
	}

	publicKeyFile, err := os.Open(abs)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}

func InitJWTAuthentication() *JWTAuthentication {
	if jwtAuthentication == nil {
		jwtAuthentication = &JWTAuthentication{
			privateKey: getPrivateKey(),
			PublicKey:  getPublicKey(),
		}
	}

	return jwtAuthentication
}

func (authentication *JWTAuthentication) GenerateToken(userUUID string) (string, error) {
	const JWTExpirationDelta = 72
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * time.Duration(JWTExpirationDelta)).Unix(),
		"iat": time.Now().Unix(),
		"sub": userUUID,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(authentication.privateKey)
	if err != nil {
		panic(err)
		return "", err
	}
	fmt.Println(tokenString, err)
	return tokenString, nil
}

func (authentication *JWTAuthentication) Authenticate(user User) bool {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testing"), 10)
	uuidStr, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	testUser := User{
		UUID:     uuidStr.String(),
		Username: "John Doe",
		Password: string(hashedPassword),
	}

	return user.Username == testUser.Username && bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(user.Password)) == nil
}

func (authentication *JWTAuthentication) getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainder := tm.Sub(time.Now())
		if remainder > 0 {
			return int(remainder.Seconds() + expireOffset)
		}
	}
	return expireOffset
}
