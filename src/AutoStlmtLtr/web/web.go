package web

import (
	. "../config"
	. "../utilities"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var LocalFile = make([]string, 0)

func GetToken() string {
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
	//Log.Println(data)
	//Log.Println(data["access_token"])
	return data["access_token"]
}

func GetPdfSyn(fileName string, paths string, token string, docType string, fb string) {
	Log.Println(fb)
	//	Log.Println(pathtofile)
	Log.Println("getPdfSyn call to synergize fileName:" + fileName)
	name := strings.TrimRight(strings.SplitAfter(fileName, ".")[0], ".")

	fileName = ".\\tmp_images\\" + name + ".pdf"
	//Log.Println("Src file = " + pathtofile+"\\"+fileName)
	Log.Println(fileName)

	// get the pdf
	url := JobConf.PdfUrl + fb + "/" + docType + "/pdf?username=" + JobConf.PdfUid + "&password=" + Conf.PdfPwd
	Log.Println(url)

	request, _ := http.NewRequest("GET", url, nil)
	//request.Header.Set("Content-Type", "application/pdf")
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		Log.Printf("The HTTP request failed with error %s\n", err)
	}
	if response.StatusCode != http.StatusOK {
		Log.Println("Status error: %v", response.StatusCode)
	} else {
		img, err := os.Create(fileName)
		if err != nil {
			Log.Println(err)
		}
		defer img.Close()

		//data, _ := ioutil.ReadAll(response.Body)
		b, _ := io.Copy(img, response.Body)
		if err != nil {
			Log.Println(err)
		}
		Log.Println("File Size: ", b)
		img.Close()
		LocalFile = append(LocalFile, fileName)
	}
}
