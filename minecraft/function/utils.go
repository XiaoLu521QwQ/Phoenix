package function

import (
"crypto/sha256"
"encoding/hex"
"io"
"os"
"strconv"
)

func SliceAtoi(sa []string) ([]float64, error) {
	si := make([]float64, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.ParseFloat(a, 64)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}

func GetHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func AddVector(v1, v2 Vector) Vector {
	v1[0] += v2[0]
	v1[1] += v2[1]
	v1[2] += v2[2]
	return v1
}