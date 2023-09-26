package utils

import (
	"encoding/binary"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func GenerateUniqueID() uint32 {

	time := time.Now().UnixNano()

	randomByte := make([]byte, 4)
	_, err := rand.Read(randomByte)

	if err != nil {
		log.Fatal(err)
	}

	return uint32(time) + binary.BigEndian.Uint32(randomByte)
}

func GenerateRandomString(length int) string {
	charSet := "qwertyuioplkjhgfdsazxcvbnmQWERTYUIOPLKJHGFDSAZXCVBNM0987654321"

	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(result)
}

func CheckOrder(num []byte) bool {
	sum := 0
	lenDigit := len(num)
	parity := lenDigit / 2

	for i := 0; i < lenDigit-1; i++ {
		digit := int(num[i])
		if i%2 == parity {
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

func ConvertByte2Int(numOrder []byte) (int, error) {
	numStr := string(numOrder)

	numberInt, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}
	return numberInt, nil
}
