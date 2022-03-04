package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func GetText(fn string) (string, error) {
	if fn == "" {
		return "", errors.New("no input file specified")
	}
	if fn == "-" {
		log.Printf("STDIN mode, enter text plaese")
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", fn, err)
	}
	return string(b), nil
}
