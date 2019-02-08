package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Claim struct {
	CLAIM_ID         int `json:"CLAIM_ID"`
	DETAIL_LINE_ID   int `json:"DETAIL_LINE_ID"`
	BILL_NUMBER      string `json:"BILL_NUMBER"`
	CLAIM_STATUS     string `json:"CLAIM_STATUS"`
	USER3            string `json:"USER3"`
	CONTACT_NAME     string `json:"CONTACT_NAME"`
	CLAIMANT_EMAIL   string `json:"CLAIMANT_EMAIL"`
	ASSIGNED_TO      string `json:"ASSIGNED_TO"`
	ITEM_DESCR       string `json:"ITEM_DESCR"`
	ITEM_IS_REQUIRED string `json:"ITEM_IS_REQUIRED"`
}

type Note struct {
	THE_NOTE string `json:"THE_NOTE"`
}

var err error

func main() {

	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", "eric123456789123")

	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	claims := make([]Claim,0)
	response, err := http.Get("http://localhost:1340/delDocStlmt")
	if err != nil {
		fmt.Printf("The http request failed with error %s\n", err)
	}else{
		data, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(data, &claims)
		fmt.Println(claims[1].CLAIM_ID)
	}

	for i:=0; i < len(claims); i++ {
		fmt.Println( strconv.Itoa(claims[i].CLAIM_ID) + " " +  claims[i].BILL_NUMBER)
		sendEmail(s, claims[i].CLAIM_ID, claims[i].BILL_NUMBER, claims[i].CONTACT_NAME, claims[i].USER3, claims[i].DETAIL_LINE_ID)
	}
}

func sendEmail(s gomail.Sender,  CLAIM_ID int, BILL string, CONTACT string, Assigned string, DLID int ){
	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", "etanner@dylt.com")
	m.SetHeader("To", "etanner@dylt.com") //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")

	notes := make([]Note,0)
	response, err := http.Get("http://localhost:1340/delDocNote?billNo="+ strconv.Itoa(DLID))
	if err != nil {
		fmt.Printf("The http request failed with error %s\n", err)
	}else{
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("The ReadAll failed with error %s\n", err)
		}else {
			json.Unmarshal(data, &notes)
			fmt.Println(notes[0].THE_NOTE)
		}
	}

	var str= "Settlement Letter, Claim ID " + strconv.Itoa(CLAIM_ID) + ", for Pro " + BILL
	//var msg = `Dear `
//	log.Println(str)

	m.SetHeader("Subject", str)
//m.SetBody("text/html" "Dear <b>" + CONTACT + ",</b>" + response.Body.

	m.SetBody("text/html", "Dear <b>" + CONTACT + ",</b>"+
	"<br><br>"+
		notes[0].THE_NOTE +
	"<br><br>"+
	"Closing: Sincerely,<br>"+
	"         " + Assigned +  " <br>"+
	"         Claims Analyst <br><br>"+
	"Attachment: Delivery Receipt for the same Freight Bill #" + BILL + " from Synergize")

	m.Attach("Z:\\1\\427\\429\\Syn17163028.tif")

	if err := gomail.Send(s, m); err != nil {
		log.Printf("Could not send email to %q: %v", CONTACT, err)
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(15 * time.Second)
	<- timer1.C


	m.Reset()
}
