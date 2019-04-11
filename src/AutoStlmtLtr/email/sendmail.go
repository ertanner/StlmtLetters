package email

import (
	. "../config"
	. "../utilities"
	. "../web"
	"github.com/vjeantet/jodaTime"
	"gopkg.in/gomail.v2"
	"log"
	"strconv"
	//. "../db"

	"strings"
	"time"
)

func LogEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note string, multiple int, state string) {
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

func SendMailStatus() {
	//
	d := gomail.NewDialer(JobConf.HostMail, JobConf.MailPort, JobConf.MailUid, Conf.MailPwd)
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
		bodyStr = bodyStr + i.State + " - " + i.Claim_id + " assigned to " + i.Analyst + " was processed. <br> \n"
	}
	m.SetBody("text/html", bodyStr)

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email %v", err)
	}

}

func SendEmail(id, fromEmail, toEmail string, claimNo int, isValid bool) bool {

	d := gomail.NewDialer(JobConf.HostMail, JobConf.MailPort, JobConf.MailUid, Conf.MailPwd)
	S, err := d.Dial()
	if err != nil {
		Log.Fatal(err)
	}
	//set up the email
	m := gomail.NewMessage()

	// set the toEmail if error
	toEmail = "TST_Recipient@DYLT.com" //"claimsError@dylt.com"

	var subject = ""
	Log.Println("isValid status in sendmail: " + strconv.FormatBool(isValid))
	Log.Println("Claim error status in sendmail: " + strconv.FormatBool(Clm[claimNo].FileError))
	if !isValid || Clm[claimNo].FileError {
		//Error !
		Log.Println("isValid state: " + strconv.FormatBool(isValid))
		Log.Println("FileErrir state: " + strconv.FormatBool(Clm[claimNo].FileError))
		//errorStr = "Error"
		toEmail = "TST_Review@DYLT.com" //"claimsError@dylt.com"
		LogEmail(id, fromEmail, toEmail, strconv.Itoa(Clm[claimNo].Detail_line_id), strconv.Itoa(Clm[claimNo].Claim_id),
			Clm[claimNo].Bill_number, Clm[claimNo].Contact_name, Clm[claimNo].Assigned_to, Clm[claimNo].Note, Clm[claimNo].Multiple, "Error")
		Log.Println("Error with formating.  Sending to error email box.")
		m.SetHeader("From", fromEmail)
		m.SetHeader("To", toEmail) //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
		subject = "ERROR !!! - Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number + " - Error !!!"
	} else {
		// No Error
		LogEmail(id, fromEmail, toEmail, strconv.Itoa(Clm[claimNo].Detail_line_id), strconv.Itoa(Clm[claimNo].Claim_id),
			Clm[claimNo].Bill_number, Clm[claimNo].Contact_name, Clm[claimNo].Assigned_to, Clm[claimNo].Note, Clm[claimNo].Multiple, "Success")
		Log.Println("Formatting is correct.  Sending out email.")
		m.SetHeader("From", fromEmail)
		m.SetHeader("To", toEmail)                        //, "claims@madegoods.com") //Clm.CLAIMANT_EMAIL
		m.SetAddressHeader("Cc", "TST_CC@DYLT.com", "")   // Claims@dylt.com
		m.SetAddressHeader("Bcc", "TST_BCC@DYLT.com", "") // Clm[claimNo].SalesRep+ "@dylt.com", Clm[claimNo].SalesRep )
		subject = "Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number
	}

	// set up the email with the data from above
	//var subject = ""
	//if errorStr > "" {
	//	subject = "ERROR !!! - Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number + " - Error !!!"
	//} else {
	//	subject = "Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number
	//}

	m.SetHeader("Subject", subject)

	a := strings.TrimPrefix(OFiles[0], ".\\tmp_images\\")
	Log.Println("File Name Send Email " + a)

	if len(LocalFile) > 0 {
		//m.Embed(LocalFile[0])
		Log.Println("Attaching LocalFile")
		for i := 0; i < len(LocalFile); i++ {
			Log.Println(LocalFile[i])
			m.Attach(LocalFile[i])
		}
	}
	//if len(DocsNeeded) > 0 {
	//	Log.Println("Attaching DocsNeeded")
	//	for l := 0; l < len(DocsNeeded); l++ {
	//		Log.Println(DocsNeeded[l])
	//		m.Attach(DocsNeeded[l])
	//	}
	//}

	headerStr := GetHeadder(claimNo)
	//Log.Println(headerStr)
	footerStr := GetFooter(claimNo)
	//Log.Println(footerStr)

	m.Embed(".\\tmp_images\\logo.tif")

	var bodyStr = `<img src="cid:logo.tif" alt="My image" /><br><br><br>` + headerStr + Clm[claimNo].Note + footerStr
	//Log.Println("note from email - " + Clm[claimNo].Note)

	m.SetBody("text/html", bodyStr)

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		Log.Printf("Could not send email to %q: %v", Clm[claimNo].Contact_name, err)
		return true
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(15 * time.Second)
	<-timer1.C

	m.Reset()
	return false
}
