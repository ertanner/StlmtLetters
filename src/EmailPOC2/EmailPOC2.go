package main

import "C"
import (
	"bytes"
	"database/sql"
	_ "github.com/alexbrainman/odbc"
	"github.com/martoche/pdf"
	"gopkg.in/gomail.v2"

	"io"
	"strings"
	//"gopkg.in/gomail.v2"
	"log"
	"time"

	"encoding/json"
	"fmt"

	"os"
	"strconv"
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
	Note             string
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
		fmt.Println("Ping error")
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
		fmt.Println("Ping error ", err)
	} else {
		log.Println("The Synergize database connection is opened...")
	}
}
func main() {
	fmt.Println("Starting mail program.")

	// clean up tmp files
	clearTmpFiles()

	// get base config
	Conf = LoadConfiguration("acxpuwd")
	fmt.Println("Config = " + Conf.MailPwd)

	// connect to the db to get the DB config
	GetConfig()
	fmt.Println("Got DB config")

	if err != nil {
		panic(err)
	}

	for i := 0; i < len(Mc); i++ {
		GetOFiles()
		GetClaims(strconv.Itoa(Mc[i].Id))
		fmt.Println("Got Claims")

		for j := 0; j < len(Clm); j++ {
			GetNote(strconv.Itoa(Clm[j].Detail_line_id))
			fmt.Println("Got Notes")

			for k := 0; k < len(Clm); k++ {
				GetFilesLoc(Clm[k].Bill_number)
				fmt.Println("Got Files")
			}
			//
			// send email
			sendEmail(Mc[i].Email_from, Mc[i].Email_to, strconv.Itoa(Clm[j].Detail_line_id), strconv.Itoa(Clm[j].Claim_id), Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Note)
			//
		}
	}

	// clean up tmp files
	//clearTmpFiles()

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
		//fmt.Println(cl.Id, cl.Idname, cl.Docsneeded, cl.Email_from, cl.Email_to, cl.Subject, cl.Greeting, cl.Body, cl.Closing, cl.Note)
		//MC = append(MC, MailConfig{id, idname, docsneeded, email_from, email_to, subject, greeting, body, closing, note})
	}
	Mc = mc
}

func GetClaims(listId string) {
	fmt.Println(listId)
	// fetch mail config data from database
	stmt := `select C.CLAIM_ID,
    t.DETAIL_LINE_ID, t.BILL_NUMBER, 
    C.CLAIM_STATUS, C.user3, c.CONTACT_NAME, c.CLAIMANT_EMAIL,
    coalesce(LC.ASSIGNED_TO, 'nil'), 
    coalesce(LI.ITEM_DESCR, 'nil'), 
    coalesce(LI.ITEM_IS_REQUIRED, 'nil')
	from CLAIM C 
	join LIST_CHECKIN LC on lc.LIST_CODE = c.CLAIM_ID
	join LIST_ITEM LI on lc.LIST_ID = li.ITEM_ID
	join TLORDER T on C.ORDER_ID = t.DETAIL_LINE_ID
	join CLIENT CL on T.CUSTOMER = CL.CLIENT_ID
	join DYLT_SETTLEMENT_LETTERS sl on lc.LIST_ID = sl.ID and sl.END_DATE is null
	where lc.UPDATED_WHEN > current date - 7 days
	  and lc.List_id = ` + listId + ` 
	and C.CLAIM_STATUS in ('CLOSED', 'OPEN')
	and lc.IS_COMPLETE = 'True'
	with ur`

	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		claim := new(Claim)
		err := rows.Scan(&claim.Claim_id, &claim.Detail_line_id, &claim.Bill_number, &claim.Claim_status, &claim.User3, &claim.Contact_name, &claim.Claimant_email, &claim.Assigned_to, &claim.Item_descr, &claim.Item_is_required)
		if err != nil {
			log.Fatal(err)
		}
		Clm = append(Clm, claim)
		fmt.Println(claim)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func GetNote(dlid string) {
	for i := 0; i < len(Clm); i++ {
		stmt := `SELECT coalesce(cast(THE_NOTE as varchar(32000)), '')  FROM NOTES N WHERE PROG_TABLE = 'TLORDER'  AND NOTE_TYPE = '3'  AND ID_KEY = '` + dlid + `'`
		//fmt.Println(stmt)
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
			//fmt.Println(claim.Note)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}
func GetFilesLoc(billNumber string) {
	// generate the string of comma delimited values to put into the frtBill values
	stmt := "select c.FBNumber, c.Bill_To_Code, c.In_DocFamilyID, m.In_DocLocation, m.In_DocID, m.In_DocFileExt, m.In_DocCreated, m.In_DocTypeID \n" +
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
	for rows.Next() {
		f := new(File)
		err = rows.Scan(&f.FBNumber, &f.BillToCode, &f.DocName, &f.DocLocation, &f.DocId, &f.DocExt, &f.DocCreated, &f.DocTypeID)
		if err != nil {
			log.Println(err)
		}
		Files = append(Files, f)
		fileName := f.DocId + "." + f.DocExt
		//Get the file
		PrepareFile(fileName, f.DocLocation)
	}
}
func GetOFiles() {

	// copy NMFTA file first as a test
	pathtofile := "O:\\DEPARTMENTS\\Claims\\_PDF\\"
	fileName := "NMFTA_Item_300105.pdf"
	copyFile(pathtofile+"\\"+fileName, ".\\tmp_images\\"+fileName)
}

func PrepareFile(fileName string, paths string) {
	// pull each file for a specific FBNumber into a local dir and the process it.
	var pathtofile string

	//	fmt.Println("File path and name: " + fileName)
	pathtofile = strings.TrimPrefix(paths, "1\\DELIVERYDOCS\\")
	pathtofile = "Z:\\" + pathtofile

	//	fmt.Println(pathtofile)
	fmt.Println(".\\tmp_images\\" + fileName)
	copyFile(pathtofile+"\\"+fileName, ".\\tmp_images\\"+fileName)

	//ImageConv(pathtofile + "\\" + fileName)
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

func imageConv() {
	r, err := pdf.Open("strings.pdf")
	if err != nil {
		fmt.Println(err)
	}

	p, err := r.GetPlainText()
	if err != nil {
		fmt.Println(err)
	}

	buf, ok := p.(*bytes.Buffer)
	if !ok {
		fmt.Println("the library no longer uses bytes.Buffer to implement io.Reader")
	}

	fmt.Println(buf.String())
}

func sendEmail(fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note string) {

	fmt.Println("Config = " + Conf.MailPwd)
	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Conf.MailPwd)

	S, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}

	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", toEmail) //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")

	//get the delivery doc's from Synergize
	var pathtofile string
	var fileName string
	fmt.Println("Z:\\" + pathtofile)
	fmt.Println(fileName)

	// set up the email with the data from above
	var str = "Settlement Letter, Claim ID " + claimID + ", for Pro " + billNo
	m.SetHeader("Subject", str)

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
		"Attachment: Delivery Receipt for the same Freight Bill #" + billNo + " from Synergize"

	m.SetBody("text/html", bodyStr)

	//m.Attach("Z:\\" + pathtofile + "\\" + fileName)

	m.Embed(fileName)

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email to %q: %v", contact, err)
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(15 * time.Second)
	<-timer1.C

	m.Reset()
}

func LoadConfiguration(file string) Config {
	var config Config
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	fmt.Println("File = " + pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	fmt.Println("M" + config.MailPwd)
	fmt.Println("S" + config.SSPwd)
	return config
}

func clearTmpFiles() {
	// remove all temp files, if any
	directory := "./tmp_Images"
	fmt.Println(directory)
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	for index := range dirFiles {
		fileHere := dirFiles[index]
		nameHere := fileHere.Name()
		fullPath := directory + "\\" + nameHere
		fmt.Println(fullPath)
		os.Remove(fullPath)
	}
}
