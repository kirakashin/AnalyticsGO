package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize"
)

type Data struct {
	ViewerId                 string   `json:"viewerId"`
	Name                     string   `json:"name"`
	LastName                 string   `json:"lastName"`
	IsChatName               bool     `json:"isChatName"`
	Email                    string   `json:"email"`
	IsChatEmail              bool     `json:"isChatEmail"`
	JoinTime                 string   `json:"joinTime"`
	LeaveTime                string   `json:"leaveTime"`
	SpentTime                int      `json:"spentTime"`
	SpentTimeDeltaPercent    int      `json:"spentTimeDeltaPercent"`
	ChatCommentsTotal        int      `json:"chatCommentsTotal"`
	ChatCommentsDeltaPercent int      `json:"chatCommentsDeltaPercent"`
	AnotherFields            []string `json:"anotherFields"`
	BrowserClientInfo        *Client  `json:"browserClientInfo"`
}

type Client struct {
	UserIP                string `json:"userIP"`
	Platform              string `json:"platform"`
	BrowserClient         string `json:"browserClient"`
	ScreenData_viewPort   string `json:"screenData_viewPort"`
	ScreenData_resolution string `json:"screenData_resolution"`
}

type Status struct {
	Status string `json:"status"`
}

type Count struct {
	Count int `json:"count"`
}

var DATA []Data

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Status{Status: "ok"})
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Count{Count: len(DATA)})
}

func collectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("collecting...")
	w.Header().Set("Content-Type", "application/json")
	var data Data
	_ = json.NewDecoder(r.Body).Decode(&data)
	DATA = append(DATA, data)
	json.NewEncoder(w).Encode(DATA)
}

func reportOSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("OS.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	for i, v := range DATA {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), v.BrowserClientInfo.Platform)
	}
	if err := f.SaveAs("OS.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "OS.XLSX")
}

func reportResHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Res.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	for i, v := range DATA {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), v.BrowserClientInfo.ScreenData_resolution)
	}
	if err := f.SaveAs("Res.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "Res.XLSX")
}

type IP struct {
	Status  string
	Country string
	Org     string
}

func reportCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Country.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	var t IP
	for i, _ := range DATA {
		res, _ := http.Get("http://ip-api.com/json/" + DATA[i].BrowserClientInfo.UserIP + "?fields=17409")
		ip, _ := ioutil.ReadAll(res.Body)
		err := json.Unmarshal(ip, &t)
		if err != nil {
			fmt.Println(err)
		}
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), t.Country)
	}
	if err := f.SaveAs("Country.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "Country.XLSX")
}

func reportProviderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Provider.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	for i, _ := range DATA {
		res, _ := http.Get("http://ip-api.com/json/" + DATA[i].BrowserClientInfo.UserIP + "?fields=17409")
		var t IP
		ip, _ := ioutil.ReadAll(res.Body)
		err := json.Unmarshal(ip, &t)
		if err != nil {
			fmt.Println(err)
		}
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), t.Org)
	}
	if err := f.SaveAs("Provider.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "Provider.XLSX")
}

type Port struct {
	Port string
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ping", pingHandler).Methods("GET")
	r.HandleFunc("/stat", statHandler).Methods("GET")
	r.HandleFunc("/collect", collectHandler).Methods("POST")
	r.HandleFunc("/report_os", reportOSHandler).Methods("GET")
	r.HandleFunc("/report_res", reportResHandler).Methods("GET")
	r.HandleFunc("/report_country", reportCountryHandler).Methods("GET")
	r.HandleFunc("/report_provider", reportProviderHandler).Methods("GET")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	var port Port
	byte, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byte, &port)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Running on" + port.Port)
	log.Fatal(http.ListenAndServe(port.Port, r))
}
