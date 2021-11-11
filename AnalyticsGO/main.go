package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
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

// func collectHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("collecting data")
// 	w.Header().Set("Content-Type", "application/json")
// 	var data Data
// 	_ = json.NewDecoder(r.Body).Decode(&data)
// 	DATA = append(DATA, data)
// 	json.NewEncoder(w).Encode(DATA)
// }

func collectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("collecting data")
	w.Header().Set("Content-Type", "application/json")
	var data []Data
	_ = json.NewDecoder(r.Body).Decode(&data)
	DATA = append(DATA, data...)
	// fmt.Println(DATA)
	// json.NewEncoder(w).Encode(DATA)
}

func reportOSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("OS.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	f.NewSheet("All stat")
	f.DeleteSheet("Sheet1")
	statPlatform := make(map[string]int)
	statBrowser := make(map[string]int)
	f.SetCellValue("All stat", "A1", "Platform")
	f.SetCellValue("All stat", "B1", "Browser")
	for i, v := range DATA {
		f.SetCellValue("All stat", "A"+strconv.Itoa(i+2), v.BrowserClientInfo.Platform)
		statPlatform[v.BrowserClientInfo.Platform]++
		f.SetCellValue("All stat", "B"+strconv.Itoa(i+2), v.BrowserClientInfo.BrowserClient)
		statBrowser[v.BrowserClientInfo.BrowserClient]++
	}
	f.NewSheet("Stat Platform")
	f.SetCellValue("Stat Platform", "A1", "Platform")
	f.SetCellValue("Stat Platform", "B1", "Count")
	counter := 1
	for i, v := range statPlatform {
		f.SetCellValue("Stat Platform", "A"+strconv.Itoa(counter), i)
		f.SetCellValue("Stat Platform", "B"+strconv.Itoa(counter), v)
		counter++
	}
	f.NewSheet("Stat Browser")
	f.SetCellValue("Stat Browser", "A1", "Browser")
	f.SetCellValue("Stat Browser", "B1", "Count")
	counter = 1
	for i, v := range statBrowser {
		f.SetCellValue("Stat Browser", "A"+strconv.Itoa(counter), i)
		f.SetCellValue("Stat Browser", "B"+strconv.Itoa(counter), v)
		counter++
	}
	if err := f.SaveAs("OS.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "OS.xlsx")
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
	http.ServeFile(w, r, "Res.xlsx")
}

type IP struct {
	Status  string
	Country string
	ISP     string
}

func reportCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Country.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	var t IP
	for i := range DATA {
		res, _ := http.Get("http://ip-api.com/json/" + DATA[i].BrowserClientInfo.UserIP + "?fields=16897")
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
	http.ServeFile(w, r, "Country.xlsx")
}

func reportProviderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Provider.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	for i := range DATA {
		res, _ := http.Get("http://ip-api.com/json/" + DATA[i].BrowserClientInfo.UserIP + "?fields=16897")
		var t IP
		ip, _ := ioutil.ReadAll(res.Body)
		err := json.Unmarshal(ip, &t)
		if err != nil {
			fmt.Println(err)
		}
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), t.ISP)
	}
	if err := f.SaveAs("Provider.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "Provider.xlsx")
}

func ipToCP(ipip []string, c chan []IP) {
	var ips []IP

	x, _ := json.Marshal(ipip)
	re := bytes.NewReader(x)
	res, _ := http.Post("http://ip-api.com/batch?fields=16897", "application/json", re)
	ip, _ := ioutil.ReadAll(res.Body)
	err := json.Unmarshal(ip, &ips)
	if err != nil {
		fmt.Println(err)
	}

	c <- ips
}

func reportCPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("CP.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	f.NewSheet("All stat")
	f.DeleteSheet("Sheet1")
	var ips []IP
	var ipip []string
	for _, v := range DATA {
		ipip = append(ipip, v.BrowserClientInfo.UserIP)
	}
	ALERT := 1
	index := 0
	c := make(chan []IP)
	for ALERT <= 15 {
		if index+100 >= len(ipip) {
			go ipToCP(ipip[index:], c)
			break
		} else {
			go ipToCP(ipip[index:index+100], c)
		}
		index += 100
		ALERT++
	}
	if ALERT == 16 {
		ALERT--
	}
	for ALERT > 0 {
		ips = append(ips, <-c...)
		ALERT--
	}

	for i, v := range ips {
		f.SetCellValue("All stat", "A"+strconv.Itoa(i+1), v.Country)
		f.SetCellValue("All stat", "B"+strconv.Itoa(i+1), v.ISP)
	}
	if err := f.SaveAs("CP.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "CP.xlsx")
}

type stamp struct {
	time  string
	start bool
	count int
}

type stampSlice []stamp

func (p stampSlice) Len() int {
	return len(p)
}

func (p stampSlice) Less(i, j int) bool {
	return p[i].time < p[j].time
}

func (p stampSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func reportPeaksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("Peaks.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	var timeStamps stampSlice
	for _, v := range DATA {
		if v.JoinTime[:5] != "0001-" && v.LeaveTime[:5] != "0001-" {
			timeStamps = append(timeStamps, stamp{v.JoinTime[:19], true, 0})
			timeStamps = append(timeStamps, stamp{v.LeaveTime[:19], false, 0})
		}
	}
	// fmt.Println(timeStamps)
	sort.Sort(timeStamps)
	// fmt.Println(timeStamps)

	counter := 0
	for i := range timeStamps {
		// fmt.Println(counter)
		// fmt.Println(v)
		if timeStamps[i].start {
			counter++
			timeStamps[i].count = counter
		} else {
			counter--
			timeStamps[i].count = counter
		}
		// fmt.Println(v)
	}
	// fmt.Println(timeStamps)
	index := 1
	for i, v := range timeStamps[:timeStamps.Len()-1] {
		if v.time != timeStamps[i+1].time {
			f.SetCellValue("Sheet1", "A"+strconv.Itoa(index), v.time)
			f.SetCellValue("Sheet1", "B"+strconv.Itoa(index), timeStamps[i+1].time)
			f.SetCellValue("Sheet1", "C"+strconv.Itoa(index), v.count)
			index++
		}
	}
	if err := f.SaveAs("Peaks.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "Peaks.xlsx")
}

// func reportAllHandler(w http.ResponseWriter, r *http.Request) {

// }

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
	r.HandleFunc("/report_cp", reportCPHandler).Methods("GET")

	r.HandleFunc("/report_peaks", reportPeaksHandler).Methods("GET")
	// r.HandleFunc("/report_all", reportAllHandler).Methods("GET")

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
