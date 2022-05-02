package utils

import "fmt"

func HandleError(err error) {
	fmt.Printf("Error: %+v\n", err)
}
