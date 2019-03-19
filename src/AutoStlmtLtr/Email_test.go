package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func init() {
	//

}
func TestGetOFiles(t *testing.T) {
	fmt.Println("TestGetOFiles")
	clearTmpFiles()
	GetOFiles()
	_, err := os.Stat("NMFTA_Item_300105.pdf")
	if os.IsNotExist(err) {
		assert.FileExists(t, "./tmp_Images/NMFTA_Item_300105.pdf", "Error")
	}
	clearTmpFiles()
}

func Test_clearTmpFiles(t *testing.T) {
	fmt.Println("Test_clearTmpFiles")
	clearTmpFiles()
	assert.NotEmpty(t, "./tmp_Images/")

}
