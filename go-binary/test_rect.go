package main

import (
	"fmt"

	"github.com/signintech/gopdf"
)

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	
	// Test Rectangle signature
	// If this compiles, we know the signature
	pdf.Rectangle(10, 10, 50, 50, "D", 0, 0)
	
	err := pdf.WritePdf("test_rect.pdf")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success")
	}
}
