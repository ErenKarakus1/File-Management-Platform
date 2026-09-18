package password

import "golang.org/x/crypto/bcrypt"

func GenerateHash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CompareHashAndPassword(hash, value string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
}
