package main

import "C"
import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"github.com/vjeantet/jodaTime"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type MailConfig struct {
	Id         int
	Idname     string
	Docsneeded string
	Email_from string
	Email_to   string
	Subject    string
	Greeting   string
	Body       string
	Closing    string
	Note       string
}

type Claim struct {
	LcId             int
	Claim_id         int
	Detail_line_id   int
	Bill_number      string
	Claim_status     string
	User3            string
	Contact_name     string
	Claimant_email   string
	Assigned_to      string
	Item_descr       string
	Item_is_required string
	POD_Signed       string
	POD_Date         string
	Note             string
	Multiple         int
}

type File struct {
	FBNumber    string
	BillToCode  string
	DocName     string
	DocLocation string
	DocId       string
	DocExt      string
	DocCreated  string
	DocTypeID   int
	DocType     string
}
type State struct {
	Claim_id         string
	Detail_line_id   string
	Bill_number      string
	Claim_status     string
	Email_from       string
	Email_to         string
	User3            string
	Contact_name     string
	Claimant_email   string
	Assigned_to      string
	Item_descr       string
	Item_is_required string
	Note             string
	Multiple         string
	State            string
}
type Config struct {
	MailPwd string `json:"MailPwd"`
	SSPwd   string `json:"SSPwd"`
}

var err error
var Conf Config
var Clm = make([]*Claim, 0)
var Mc = make([]*MailConfig, 0)
var Files = make([]*File, 0)
var LocalFile = make([]string, 0)
var OFiles = make([]string, 0)
var DocsNeeded = make([]string, 0)
var MailStatus = make([]*State, 0)

var buf bytes.Buffer
var logger = log.New(&buf, "logger: ", log.Lshortfile)

var TmwDb *sql.DB
var SynDb *sql.DB

func init() {

	//
	TmwDb, err = sql.Open("odbc", "DSN=DYLT_REP")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to TMW database...")

	TmwDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	TmwDb.SetConnMaxLifetime(30 * time.Minute)
	TmwDb.SetMaxIdleConns(0)

	err = TmwDb.Ping()
	if err != nil {
		log.Println("Ping error")
	} else {
		log.Println("Database connection to TMW is opened...")
	}

	// connect to the Synergize db
	SynDb, err = sql.Open("odbc",
		"DSN=SYNERGIZE;UID=sa;PWD=Syn3rg1ze")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to the Synergize database...")
	SynDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	SynDb.SetConnMaxLifetime(30 * time.Minute)
	SynDb.SetMaxIdleConns(0)

	err = SynDb.Ping()
	if err != nil {
		log.Println("Ping error ", err)
	} else {
		log.Println("The Synergize database connection is opened...")
	}
}

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("log_file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}

	//defer to close when you're done with it
	defer f.Close()
	//set output of logs to f
	log.SetOutput(f)
	log.Println("Starting mail program.")

	// clean up tmp files from prior run
	clearTmpFiles()

	// get base config
	Conf = LoadConfiguration("acxpuwd")

	// connect to the db to get the DB config showing the checklist items to be processed
	GetConfig()
	log.Println("Got DB config")

	if err != nil {
		panic(err)
	}

	for i := 0; i < len(Mc); i++ {
		// get the files off the O drive
		GetOFiles()

		ParseDocs(Mc[i].Docsneeded)
		GetClaims(strconv.Itoa(Mc[i].Id))
		log.Println("Got Claims")

		for j := 0; j < len(Clm); j++ {

			log.Println("Get Notes")
			GetNote(strconv.Itoa(Clm[j].Detail_line_id))

			log.Println("Get multiple list items for one claim")

			log.Println("Parsed Email Config")

			GetFilesLoc(Clm[j].Bill_number)
			log.Println("Got Files")
			log.Println(Clm[j].LcId, strconv.Itoa(Mc[i].Id), Mc[i].Email_from, Mc[i].Email_to, strconv.Itoa(Clm[j].Detail_line_id), strconv.Itoa(Clm[j].Claim_id),
				Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Note, strconv.Itoa(Clm[j].Multiple))
			// send email
			if Mc[i].Id == 101 || Mc[i].Id == 146 {
				Clm[j].Note = formatNote(i, Mc[i].Id)
			}
			sendEmail(strconv.Itoa(Mc[i].Id), Mc[i].Email_from, Mc[i].Email_to, strconv.Itoa(Clm[j].Detail_line_id),
				strconv.Itoa(Clm[j].Claim_id), Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Note, Clm[j].Multiple)
			LocalFile = LocalFile[:0]
		}
		// reset Claim slice back to 0 elements
		Clm = Clm[:0]
	}

	// clean up tmp files
	sendMailStatus()
	clearTmpFiles()
}

func GetConfig() {
	//	fetch mail config data from database

	stmt := `select distinct ID, IDNAME, DOCSNEEDED, EMAIL_FROM, EMAIL_TO, SUBJECT, GREETING, BODY, CLOSING, NOTE 
			 from DYLT_SETTLEMENT_LETTERS 
			 where END_DATE is null
			 with ur`
	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	mc := make([]*MailConfig, 0)
	for rows.Next() {
		cl := new(MailConfig)
		err := rows.Scan(&cl.Id, &cl.Idname, &cl.Docsneeded, &cl.Email_from, &cl.Email_to, &cl.Subject, &cl.Greeting, &cl.Body, &cl.Closing, &cl.Note)
		if err != nil {
			log.Fatal(err)
		}
		mc = append(mc, cl)
	}
	log.Println(mc)
	Mc = mc
}

func GetClaims(listId string) {
	log.Println(listId)
	// fetch mail config data from database
	stmt := `select lc.List_id, C.CLAIM_ID,
    t.DETAIL_LINE_ID, t.BILL_NUMBER, 
    C.CLAIM_STATUS, C.user3, c.CONTACT_NAME, c.CLAIMANT_EMAIL,
    coalesce(LC.ASSIGNED_TO, 'nil'), 
    coalesce(LI.ITEM_DESCR, 'nil'), 
    coalesce(LI.ITEM_IS_REQUIRED, 'nil'),
	count(sl.ID) Multiple
	from CLAIM C 
	join LIST_CHECKIN LC on lc.LIST_CODE = c.CLAIM_ID
	join LIST_ITEM LI on lc.LIST_ID = li.ITEM_ID
	join TLORDER T on C.ORDER_ID = t.DETAIL_LINE_ID
	join CLIENT CL on T.CUSTOMER = CL.CLIENT_ID
	join DYLT_SETTLEMENT_LETTERS sl on lc.LIST_ID = sl.ID and sl.END_DATE is null
	left join POD D on d.DLID = t.DETAIL_LINE_ID and d.TX_TYPE = 'Drop'
	where lc.UPDATED_WHEN > current date - 7 days
	  and lc.List_id = ` + listId + ` 
	and C.CLAIM_STATUS in ('CLOSED', 'OPEN')
	and lc.IS_COMPLETE = 'True'
	group by lc.List_id, C.CLAIM_ID, t.DETAIL_LINE_ID, t.BILL_NUMBER,C.CLAIM_STATUS, C.user3, c.CONTACT_NAME, c.CLAIMANT_EMAIL,
	coalesce(LC.ASSIGNED_TO, 'nil'), coalesce(LI.ITEM_DESCR, 'nil'), coalesce(LI.ITEM_IS_REQUIRED, 'nil')
	with ur`

	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		claim := new(Claim)
		err := rows.Scan(&claim.LcId, &claim.Claim_id, &claim.Detail_line_id, &claim.Bill_number, &claim.Claim_status,
			&claim.User3, &claim.Contact_name, &claim.Claimant_email, &claim.Assigned_to, &claim.Item_descr, &claim.Item_is_required,
			&claim.Multiple)
		if err != nil {
			log.Fatal(err)
		}
		Clm = append(Clm, claim)
		log.Println(claim)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

}

func GetNote(dlid string) {

	for i := 0; i < len(Clm); i++ {
		stmt := `SELECT coalesce(cast(THE_NOTE as varchar(32000)), '')  FROM NOTES N WHERE PROG_TABLE = 'TLORDER'  AND NOTE_TYPE = '3'  AND ID_KEY = '` + dlid + `'`
		//log.Println(stmt)
		rows, err := TmwDb.Query(stmt)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			claim := new(Claim)
			err := rows.Scan(&claim.Note)
			if err != nil {
				log.Fatal(err)
			}
			Clm[i].Note = claim.Note
			log.Println(claim.Note)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}
func GetFilesLoc(billNumber string) {
	// generate the string of comma delimited values to put into the frtBill values
	stmt := "select c.FBNumber, c.Bill_To_Code, c.In_DocFamilyID, m.In_DocLocation, m.In_DocID, m.In_DocFileExt, m.In_DocCreated, m.In_DocTypeID, \n" +
		"case \n" +
		"when m.In_DocTypeID = 25 then 'dr' \n" +
		"when m.In_DocTypeID = 16 then 'bol' \n" +
		"end as DocType \n" +
		" from DELIVERYDOCS.dbo.Child C  \n" +
		" inner join DELIVERYDOCS.dbo.Main M on M.In_DocID = C.In_DocFamilyID  \n" +
		" where FBNumber = '" + billNumber + "'  \n" +
		" and M.DeliveryDate is not null \n" +
		" and m.In_DocTypeID in (16, 25)"
	rows, err := SynDb.Query(stmt)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	LocalFile = LocalFile[:0]
	for rows.Next() {
		f := new(File)
		err = rows.Scan(&f.FBNumber, &f.BillToCode, &f.DocName, &f.DocLocation, &f.DocId, &f.DocExt, &f.DocCreated, &f.DocTypeID, &f.DocType)
		if err != nil {
			log.Println(err)
		}
		Files = append(Files, f)
		fileName := f.DocId + "." + f.DocExt
		log.Println("FileName = " + fileName)

		// get the token
		token := getToken()

		//Get the file
		PrepareFile(fileName, f.DocLocation, token, f.DocType, f.FBNumber)
	}
}
func GetOFiles() {

	// copy NMFTA file first as a test
	pathtofile := "O:\\DEPARTMENTS\\Claims\\_PDF\\"
	fileName := "NMFTA_Item_300105.pdf"
	copyFile(pathtofile+"\\"+fileName, ".\\tmp_images\\"+fileName)
	OFiles = append(OFiles, ".\\tmp_images\\"+fileName)
}

func PrepareFile(fileName string, paths string, token string, docType string, fb string) {

	//	log.Println(pathtofile)
	name := strings.TrimRight(strings.SplitAfter(fileName, ".")[0], ".")

	fileName = ".\\tmp_images\\" + name + ".pdf"
	//log.Println("Src file = " + pathtofile+"\\"+fileName)
	log.Println(fileName)
	//	copyFile(pathtofile+"\\"+fileName, ".\\tmp_images\\"+fileName)

	// get the pdf
	url := "https://api.dylt.com/image/" + fb + "/" + docType + "/pdf?userName=AUTORGH&password=alwayskeepasmile"
	//log.Println(url)

	request, _ := http.NewRequest("GET", url, nil)
	//request.Header.Set("Content-Type", "application/pdf")
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
	}
	if response.StatusCode != http.StatusOK {
		fmt.Errorf("Status error: %v", response.StatusCode)
	} else {
		img, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer img.Close()

		//data, _ := ioutil.ReadAll(response.Body)
		b, _ := io.Copy(img, response.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("File Size: ", b)
		img.Close()
		LocalFile = append(LocalFile, fileName)
	}
}

// Copy a file
func copyFile(src, dest string) {
	// Open original file
	originalFile, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer originalFile.Close()

	// Create new file
	newFile, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	// Copy the bytes to destination from source
	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Copied %d bytes.", bytesWritten)

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func ParseDocs(docsNeeded string) {
	//
	log.Println(docsNeeded)
	r := csv.NewReader(strings.NewReader(docsNeeded))
	for {
		record, err := r.Read()
		DocsNeeded = record
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
		}
	}
	log.Println(DocsNeeded)
}
func parseEmailConfig(i int, j int) {
	//

}
func imageConv(image string) {
	log.Println("Convert image: " + image)
	//
	////bimg.VipsCacheDropAll()
	//log.Println("Read iamge")
	//buffer, err := bimg.Read(image)
	//if err != nil {
	//	log.Fprintln(os.Stderr, err)
	//}
	//log.Println("convert")
	//newImage, err := bimg.NewImage(buffer).Convert(bimg.PDF)
	//if err != nil {
	//	log.Fprintln(os.Stderr, err)
	//}
	//
	//log.Println("writre")
	//if bimg.NewImage(newImage).Type() == "pdf" {
	//	log.Fprintln(os.Stderr, "The image was converted")
	//}
	//

	//buffer, err := bimg.Read(image)
	//if err != nil {
	//	log.Fprintln(os.Stderr, err)
	//	log.Fatal(err)
	//}
	//
	//newImage, err := bimg.NewImage(buffer).Rotate(90)
	//if err != nil {
	//	log.Fprintln(os.Stderr, err)
	//	log.Fatal(err)
	//}
	//log.Println("Write the new image to disk")
	//bimg.Write(image, newImage)

}

func sendEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note string, multiple int) {

	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Conf.MailPwd)

	S, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	//
	// set the toEmail if error
	//
	if multiple > 1 || claimID == "" || billNo == "" || contact == "" || assignedTo == "" || len(OFiles[0]) == 0 || note == "" {
		toEmail = "etanner@dylt.com" //"claimsError@dylt.com"
		logEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note, multiple, "Error")
		log.Println("Error with formating.  Sending to error email box.")
	} else {
		logEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note, multiple, "Success")
	}

	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", toEmail) //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")

	// set up the email with the data from above
	var str = "Settlement Letter, Claim ID " + claimID + ", for Pro " + billNo
	m.SetHeader("Subject", str)

	a := strings.TrimPrefix(OFiles[0], ".\\tmp_images\\")
	log.Println("File Name Send Email" + a)
	m.Embed(OFiles[0])

	var bodyStr = ""

	if "" == contact {
		bodyStr = "Dear <b>Dear Claimant,</b>"
	} else {
		bodyStr = "Dear <b>" + contact + ",</b>"
	}
	bodyStr = bodyStr + "<br><br>" + note +
		"<br><br>" +
		"Closing: Sincerely,<br>" +
		"         " + assignedTo + " <br>" +
		"         Claims Analyst <br><br>" +
		"Attachment: Delivery Receipt for the same Freight Bill #" + billNo + " from Synergize <br>" +
		"<img src=\"cid:" + OFiles[0] + " \" alt=\"My image\" />"

	m.SetBody("text/html", bodyStr)

	for i := 0; i < len(LocalFile); i++ {
		m.Attach(LocalFile[i])
	}

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email to %q: %v", contact, err)
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(13 * time.Second)
	<-timer1.C

	m.Reset()
}

func logEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note string, multiple int, state string) {
	// write the error
	ms := new(State)
	s := strconv.Itoa(multiple)
	ms.Claim_id = id
	ms.Email_from = fromEmail
	ms.Email_to = toEmail
	ms.Detail_line_id = dlid
	ms.Claim_id = claimID
	ms.Bill_number = billNo
	ms.Contact_name = contact
	ms.Assigned_to = assignedTo
	ms.Note = note
	ms.Multiple = s
	ms.State = state

	MailStatus = append(MailStatus, ms)

}

func formatNote(i, id int) string {
	//
	var filename = ""
	var s = ""

	if id == 101 {
		filename = "101.txt"
	}
	if id == 146 {
		filename = "146.txt"
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	s = string(b)
	return s
}

func sendMailStatus() {
	//
	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Conf.MailPwd)

	S, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", "etanner@dylt.com")
	m.SetHeader("To", "etanner@dylt.com") //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")
	date := jodaTime.Format("YYYY.MM.dd", time.Now())

	var str = "Automated Settlement Letter batch job for " + date
	m.SetHeader("Subject", str)
	m.Attach("log_file")
	var bodyStr = "<br><br>"
	for _, i := range MailStatus {
		log.Printf(i.Claim_id + "<br>")
	}
	m.SetBody("text/html", bodyStr)

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email %v", err)
	}

}
func businessDays(start string, end string) {
	//
}
func LoadConfiguration(file string) Config {
	var config Config
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		log.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func clearTmpFiles() {
	// remove all temp files, if any
	directory := "./tmp_Images"
	log.Println(directory)
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	for index := range dirFiles {
		fileHere := dirFiles[index]
		nameHere := fileHere.Name()
		fullPath := directory + "\\" + nameHere
		log.Println(fullPath)
		os.Remove(fullPath)
	}
	//os.Remove("log_file")
	date := jodaTime.Format("YYYY.MM.dd", time.Now())
	os.Rename("log_file", "log_file-"+date)
}

func getToken() string {
	//Consumer Key: x5Vxusddiy2pYqwpZytwxqkG0lW7Z6a5
	//Consumer Secret: ThzO25vxF0RDuA2U
	body := strings.NewReader(`client_secret=P0AGMIlIAFC1vEqn&grant_type=client_credentials&client_id=QLVqxgQk85apoXB8AFAeSOYTv4RR53lh`)
	req, err := http.NewRequest("POST", "https://api.dylt.com/oauth/client_credential/accesstoken?grant_type=client_credentials", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	token, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	var data map[string]string
	json.Unmarshal(token, &data)
	//log.Println(data)
	log.Println(data["access_token"])
	return data["access_token"]
}
