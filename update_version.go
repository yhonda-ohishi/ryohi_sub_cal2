package main

import (
	"fmt"
	"log"
	"github.com/your-org/ryohi-router/src/lib/dtako"
)

func main() {
	err := dtako.UpdateSwaggerDescription("docs")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Swagger version updated successfully")
}