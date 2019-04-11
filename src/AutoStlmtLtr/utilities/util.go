package utilities

import (
	. "../config"
	"encoding/csv"
	"encoding/json"
	"github.com/vjeantet/jodaTime"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// two UTF-8 functions identical except for operator comparing c to 127
func StripCtlFromUTF8(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, str)
}
func StripCtlAndExtFromUTF8(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 {
			return r
		}
		return -1
	}, str)
}
func IsLetterA(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			Log.Println("cound unicode character: " + string(r))
			return false
		}
	}
	return true
}
func ClearTmpFiles() {
	// remove all temp files, if any
	directory := "./tmp_Images"
	Log.Println(directory)
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	for index := range dirFiles {
		fileHere := dirFiles[index]
		nameHere := fileHere.Name()
		fullPath := directory + "\\" + nameHere
		Log.Println(fullPath)
		os.Remove(fullPath)
	}
	//os.Remove("log_file")
	date := jodaTime.Format("YYYY_MM_dd_HH_mm_ss", time.Now())
	os.Rename(".\\log\\log_file", "log_file-"+date)
}
func LoadConfiguration(file string) Config {
	// get pwd config
	var config Config
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		// fail the job if there is any error.
		Log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
func LoadConfigurationJob(file string) JobConfig {
	// get pwd config
	var config JobConfig
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		// fail the job if there is any error.
		Log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// Copy a file
func CopyFile(src, dest string) {
	// Open original file
	originalFile, err := os.Open(src)
	if err != nil {
		Log.Fatal(err)
	}
	defer originalFile.Close()

	// Create new file
	newFile, err := os.Create(dest)
	if err != nil {
		Log.Fatal(err)
	}
	defer newFile.Close()

	// Copy the bytes to destination from source
	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Printf("Copied %d bytes.", bytesWritten)

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		Log.Fatal(err)
	}
}
func ParseString(docsNeeded string) []string {
	//
	Log.Println("Incomming string to parseString: " + docsNeeded)
	//reset the slice to o length
	needed := make([]string, 0)

	r := csv.NewReader(strings.NewReader(docsNeeded))
	//var docs = DocsNeeded
	//fmt.Println(reflect.TypeOf(docs))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			Log.Println(err)
		}
		needed = record

		Log.Println("record1 " + record[0])
		Log.Println(DocsNeeded)
		//fmt.Println(reflect.TypeOf(record))

	}
	//DocsNeeded = append(DocsNeeded, csvData)
	Log.Println("System Docs Needed.")
	Log.Println(needed)
	return needed
}
func CopyODrive() {
	//
	pathtofiles := "O:\\DEPARTMENTS\\Claims\\_PDF\\"
	files, err := ioutil.ReadDir(pathtofiles)
	if err != nil {
		Log.Println(err)
		Log.Fatal(err)
	}
	for _, file := range files {
		CopyFile(pathtofiles+"\\"+file.Name(), ".\\tmp_images\\"+file.Name())
		OFiles = append(OFiles, ".\\tmp_images\\"+file.Name())
	}

}
func BusinessDays(from time.Time, to time.Time) int {
	//start format is '01/06/2019'
	//date := 0
	Log.Println("start date ") // + start)
	Log.Println("end date ")   // + end)

	totalDays := float32(to.Sub(from) / (24 * time.Hour))
	weekDays := float32(from.Weekday()) - float32(to.Weekday())
	businessDays := int(1 + (totalDays*5-weekDays*2)/7)
	if to.Weekday() == time.Saturday {
		businessDays--
	}
	if from.Weekday() == time.Sunday {
		businessDays--
	}

	return businessDays
}

func GetHeadder(claimNo int) string {
	str := ""
	str = "" + Clm[claimNo].Contact_name + "<br>"
	str = str + Clm[claimNo].Company + "<br>"
	if Clm[claimNo].Address1 != "" || Clm[claimNo].Address1 != "nil" {
		str = str + Clm[claimNo].Address1 + "<br> "
	} else {
		Log.Println("Address1 is null")
	}
	if Clm[claimNo].City != "" || Clm[claimNo].City != "nil" {
		str = str + Clm[claimNo].City + ", "
	} else {
		Log.Println("City is null")
	}
	if Clm[claimNo].Provence != "" || Clm[claimNo].Provence != "nil" {
		str = str + Clm[claimNo].Provence + " "
	} else {
		Log.Println("State is null")
	}
	if Clm[claimNo].PostalCode != "" || Clm[claimNo].PostalCode != "nil" {
		str = str + Clm[claimNo].PostalCode + "<br>"
	} else {
		Log.Println("State is null")
	}
	str = str + "Claimant: " + Clm[claimNo].Claimant_email + "<br>"
	str = str + "Sales: " + Clm[claimNo].SalesRep + "<br><br><br>"

	str = str + "Your Claim #: " + Clm[claimNo].TraceNo + "<br>"
	str = str + "Our Claim #: " + strconv.Itoa(Clm[claimNo].Claim_id) + "<br>"
	str = str + "Freight Bill#: " + Clm[claimNo].Bill_number + "<br>"
	str = str + "Claim Amount: $" + strconv.FormatFloat(Clm[claimNo].AmtClamed, 'f', 2, 64) + "<br><br><br>"

	//if Clm[claimNo].Contact_name == ""{
	//	str = str + "Dear Claimant, <br><br>"
	//}else{
	//	str = str + "Dear " + Clm[claimNo].Contact_name + ",<br><br>"
	//}
	return str
}
func GetFooter(claimNo int) string {
	str := "<br><br>Sincerely,<br>" +
		"         " + Clm[claimNo].Analyst + " <br>" +
		"         Claims Analyst " +
		//"<br><br>" +
		//"Attachment: Delivery Receipt for the same Freight Bill #" + Clm[claimNo].Bill_number + " from Synergize " +
		"<br><br><br><br>"

	str = str + "Note:</b>  Rebuttals must be submitted in writing with additional information supporting your claim.  "
	str = str + "Please reference claim # " + strconv.Itoa(Clm[claimNo].Claim_id) + "  and fax to 888-845-9251 or email to claims@dylt.com.<br>"

	str = str + "<br><br><br><hr><br><center>Daylight Transport LLC    1501 Hughes Way Ste 200    Long Beach  California 90810     800-468-9999     888-845-9251 Fax     www.dylt.com<center>"

	return str
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
