package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var Token = ""
var claim = new(Claim)
var mc = new(MailConfig)

func init() {
	DocsNeeded = DocsNeeded[:0]
	DocsNeeded = append(DocsNeeded, "")
}

//func TestMain(m *testing.M)  {
//	//DocsNeeded[0] = "1,2,3"
//}

func TestGetConfig(t *testing.T) {
	fmt.Println("TestGetConfig")
	GetConfig()
	if len(Mc) == 0 {
		t.Errorf("Error with TestGetConfig.")
	}
}

func TestParseDocs(t *testing.T) {
	fmt.Println("TestParseDocs")

	x := parseString("1,2")
	fmt.Println(x[0])
	if x[0] != "1" {
		t.Errorf("DocsNeeded was something other than a 1")
	}
}

func TestPrepareFile(t *testing.T) {
	fmt.Println("TestPrepareFile")
	//	PrepareFile(fileName, paths, Token, "dr", "72541394")
}

func Test_logEmail(t *testing.T) {
	fmt.Println("Test_logEmail")

	logEmail("1", "ET", "BT", "2", "3", "4", "5", "6", "7", 8, "9")
	if MailStatus[0].Claim_id != "3" {
		t.Errorf("Error with Test_logEmail.")
	}
}
func Test_formatNote(t *testing.T) {
	fmt.Println("Test_formatNote")

	claim := new(Claim)
	claim.POD_Date = time.Now()
	claim.Claim_Date = time.Now()
	claim.DaysDiff = 10
	Clm = append(Clm, claim)

	fmt.Println("Claim initalized")
	b, err := ioutil.ReadFile("101.txt")
	if err != nil {
		fmt.Print(err)
	}
	str101 := string(b)

	c, err := ioutil.ReadFile("146.txt")
	if err != nil {
		fmt.Print(err)
	}
	str146 := string(c)

	str := formatNote(0, 101)
	if !strings.ContainsAny(str, "without exception") {
		fmt.Println("inside Error 101")
		t.Errorf("Error with Test_formatNote. 101")
	}
	if strings.Compare(str101, str) <= 0 {
		t.Errorf("Error comparing 101.txt with str101 in the Test_formatNote. 101")
	}

	str = formatNote(0, 146)
	if !strings.ContainsAny(str, "Your claim") {
		fmt.Println("inside Error 146")
		t.Errorf("Error with Test_formatNote. 146")
	}
	if strings.Compare(str146, str) <= 0 {
		t.Errorf("Error comparing 146.txt with str146 in the Test_formatNote. 146")
	}
}

func Test_businessDays(t *testing.T) {
	fmt.Println("Test_businessDays")
	layoutISO := "2006-01-02"
	start := "2006-01-02"
	end := "2006-01-12"
	sd, _ := time.Parse(layoutISO, start)
	ed, _ := time.Parse(layoutISO, end)
	days := businessDays(sd, ed)
	if days < 9 || days > 9 {
		t.Errorf("Error with Test_businessDays %v", days)
	} else {
		fmt.Println("Test_businessDays ran correctly")
	}
}
func TestLoadConfiguration(t *testing.T) {
	fmt.Println("TestLoadConfiguration")
	config := LoadConfiguration("acxpuwd")
	if config.MailPwd != "judy123456789123" {
		t.Errorf("Error with config pwd!")
	} else {
		fmt.Println("TestLoadConfiguration passed.")
	}
}
func Test_clearTmpFiles(t *testing.T) {
	fmt.Println("Test_clearTmpFiles")
	clearTmpFiles()
	assert.NotEmpty(t, "./tmp_Images/")

}

func Test_getToken(t *testing.T) {
	fmt.Println("Test_getToken")
	token := getToken()
	if token == "" {
		t.Errorf("Token was empty")
	}

	url := "https://api.dylt.com/image/99493272/dr/pdf?userName=AUTORGH&password=alwayskeepasmile"
	request, _ := http.NewRequest("GET", url, nil)
	//request.Header.Set("Content-Type", "application/pdf")
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Error("Error with token")
	} else {
		fmt.Println("Token works")
	}
	Token = token
}
func Test_validateFields(t *testing.T) {
	fmt.Println("Test_validateFields")
	mc.Id = 1
	claim.Multiple = 2
	claim.Claim_id = 1
	claim.Bill_number = "1"
	claim.Contact_name = "A"
	claim.Assigned_to = "A"
	claim.Note = "A"
	claim.Analyst = "A"

	fmt.Println("Test Multiple")

	x := validateFields(*claim, *mc)
	if x {
		t.Errorf("Error with validation -  multiple")
	}

	fmt.Println("Test Bill Number")
	claim.Multiple = 0
	claim.Bill_number = ""
	x = validateFields(*claim, *mc)
	if x {
		t.Errorf("Error with validation - Bill Number")
	}

	fmt.Println("Test Contact_name")
	claim.Contact_name = ""
	claim.Bill_number = "a"
	x = validateFields(*claim, *mc)
	if x {
		t.Errorf("Error with validation - Contact_name")
	}

	fmt.Println("Test Assigned_to")
	claim.Contact_name = "a"
	claim.Assigned_to = ""
	x = validateFields(*claim, *mc)
	if x {
		t.Errorf("Error with validation - Assigned_to")
	}

	fmt.Println("Test Analyst")
	claim.Assigned_to = "a"
	claim.Analyst = "nil"
	x = validateFields(*claim, *mc)
	if x {
		t.Errorf("Error with validation - Analyst")
	}

	fmt.Println("Now test the positive")
	claim.Analyst = "A"
	x = validateFields(*claim, *mc)
	if !x {
		t.Errorf("Error with validation - Positive Test")
	}

}
func Test_copyODrive(t *testing.T) {
	fmt.Println("TestGetOFiles")
	clearTmpFiles()
	copyODrive()
	_, err := os.Stat("NMFTA_Item_300105.pdf")
	if os.IsNotExist(err) {
		assert.FileExists(t, "./tmp_Images/NMFTA_Item_300105.pdf", "Error")
	}
	clearTmpFiles()
}

func Test_sendEmail(t *testing.T) {
	fmt.Println("Test_sendEmail")
	Conf = LoadConfiguration("acxpuwd")
	//GetConfig()

	claim.Multiple = 0
	claim.Claim_id = 1
	claim.Bill_number = "1"
	claim.Contact_name = "A"
	claim.Assigned_to = "A"
	claim.Note = "A"
	claim.Analyst = "A"
	Clm = append(Clm, claim)
	isValid := true
	fmt.Println("Got past assignments")

	result := sendEmail("99", "etanner@dylt.com", "etanner@dylt.com", 0, isValid)
	if result {
		t.Errorf("Error with sending email")
	}
}
func TestGetClaims(t *testing.T) {
	fmt.Println("TestGetClaims")
	Conf = LoadConfiguration("acxpuwd")
	GetConfig()
	*endDate = "3/11/2019"
	GetClaims("99")
	fmt.Println(Clm[0].LcId)
	fmt.Println(Clm[0].Claim_id)
	fmt.Println(Clm[0].Bill_number)
	fmt.Println(Clm[0].Detail_line_id)

	fmt.Println(Clm[1].LcId)
	fmt.Println(Clm[1].Claim_id)
	fmt.Println(Clm[1].Bill_number)
	fmt.Println(Clm[1].Detail_line_id)

	if Clm[0].Claim_id != 0 {
		t.Errorf("Error with TestGetClaims")
	}
}
func TestGetNote(t *testing.T) {
	fmt.Println("TestGetNote")
	Clm[0].Bill_number = "72935851"
	Clm[0].Detail_line_id = 107997900

	GetNote("", 0)

	fmt.Println(Clm[0].Bill_number)
	fmt.Println(Clm[0].Note)
	if Clm[0].Note == "nil" {
		fmt.Println("evreything works great")
		t.Errorf("Error with TestGetNote - 'nil' ")
	}
	if Clm[0].Note == "Note" {
		t.Errorf("Error with TestGetNote")
	}
}

func TestGetFilesLoc(t *testing.T) {
	fmt.Println("TestGetFilesLoc")
	Clm[0].Bill_number = "72935851"
	Clm[0].Detail_line_id = 107997900

	GetFilesLoc(Clm[0].Bill_number)
	if Files[0].DocType != "dr" {
		fmt.Println(Files[0].DocType)
	}

}
