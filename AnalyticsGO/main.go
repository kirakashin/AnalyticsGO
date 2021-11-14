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
	"strings"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
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

var DATA []Data

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Status{Status: "ok"})
}

type Count struct {
	Count int `json:"count"`
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Count{Count: len(DATA)})
}

type Collect struct {
	Skip      int `json:"skip"`
	Collected int `json:"collected"`
}

func collectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("collecting data")
	w.Header().Set("Content-Type", "application/json")
	var data []Data
	var col Collect
	_ = json.NewDecoder(r.Body).Decode(&data)
	for _, v := range data {
		if v.BrowserClientInfo != nil {
			DATA = append(DATA, v)
			col.Collected++
		} else {
			col.Skip++
		}
	}
	json.NewEncoder(w).Encode(col)
}

type IP struct {
	Status  string
	Country string
	ISP     string
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

func reportAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("All.xlsx"))
	w.Header().Set("Content-Type", "application/octet-stream")
	f := excelize.NewFile()
	//sheets
	f.NewSheet("All_stat")
	f.DeleteSheet("Sheet1")
	f.NewSheet("DB_stat")
	f.NewSheet("Resolution_stat")
	f.NewSheet("CP_stat")
	f.NewSheet("Peaks_stat")
	//names all
	f.SetCellValue("All_stat", "A1", "viewerId")
	f.SetCellValue("All_stat", "B1", "name")
	f.SetCellValue("All_stat", "C1", "lastName")
	f.SetCellValue("All_stat", "D1", "isChatName")
	f.SetCellValue("All_stat", "E1", "email")
	f.SetCellValue("All_stat", "F1", "isChatEmail")
	f.SetCellValue("All_stat", "G1", "joinTime")
	f.SetCellValue("All_stat", "H1", "leaveTime")
	f.SetCellValue("All_stat", "I1", "spentTime")
	f.SetCellValue("All_stat", "J1", "spentTimeDeltaPercent")
	f.SetCellValue("All_stat", "K1", "chatCommentsTotal")
	f.SetCellValue("All_stat", "L1", "chatCommentsDeltaPercent")
	f.SetCellValue("All_stat", "M1", "anotherFields")
	f.SetCellValue("All_stat", "N1", "userIP")
	f.SetCellValue("All_stat", "O1", "platform")
	f.SetCellValue("All_stat", "P1", "browserClient")
	f.SetCellValue("All_stat", "Q1", "screenData_viewPort")
	f.SetCellValue("All_stat", "R1", "screenData_resolution")
	//names Device/Browser
	f.SetCellValue("DB_stat", "A1", "platform")
	f.SetCellValue("DB_stat", "B1", "browserClient")
	f.SetCellValue("DB_stat", "D1", "platform")
	f.SetCellValue("DB_stat", "E1", "COUNT")
	f.SetCellValue("DB_stat", "G1", "browserClient")
	f.SetCellValue("DB_stat", "H1", "COUNT")
	f.SetCellValue("DB_stat", "J1", "platform")
	f.SetCellValue("DB_stat", "K1", "COUNT")
	f.SetCellValue("DB_stat", "M1", "browserClient")
	f.SetCellValue("DB_stat", "N1", "COUNT")

	statPlatform := make(map[string]int)
	statBrowser := make(map[string]int)
	//names Resolution
	f.SetCellValue("Resolution_stat", "A1", "screenData_resolution")
	f.SetCellValue("Resolution_stat", "C1", "screenData_resolution")
	f.SetCellValue("Resolution_stat", "D1", "COUNT")
	f.SetCellValue("Resolution_stat", "F1", "ratio")
	f.SetCellValue("Resolution_stat", "G1", "COUNT")

	statResolution := make(map[string]int)
	//names Location/provider
	f.SetCellValue("CP_stat", "A1", "location")
	f.SetCellValue("CP_stat", "B1", "provider")
	f.SetCellValue("CP_stat", "D1", "location")
	f.SetCellValue("CP_stat", "E1", "COUNT")
	f.SetCellValue("CP_stat", "G1", "provider")
	f.SetCellValue("CP_stat", "H1", "COUNT")

	statLocation := make(map[string]int)
	statProvider := make(map[string]int)
	var ipip []string
	//names Peaks
	f.SetCellValue("Peaks_stat", "A1", "time_begin")
	f.SetCellValue("Peaks_stat", "B1", "time_end")
	f.SetCellValue("Peaks_stat", "C1", "COUNT")

	var timeStamps stampSlice
	//main loader
	for i, v := range DATA {

		index := strconv.Itoa(i + 2)
		//all loader
		f.SetCellValue("All_stat", "A"+index, v.ViewerId)
		f.SetCellValue("All_stat", "B"+index, v.Name)
		f.SetCellValue("All_stat", "C"+index, v.LastName)
		f.SetCellValue("All_stat", "D"+index, v.IsChatName)
		f.SetCellValue("All_stat", "E"+index, v.Email)
		f.SetCellValue("All_stat", "F"+index, v.IsChatEmail)
		f.SetCellValue("All_stat", "G"+index, v.JoinTime)
		f.SetCellValue("All_stat", "H"+index, v.LeaveTime)
		f.SetCellValue("All_stat", "I"+index, v.SpentTime)
		f.SetCellValue("All_stat", "J"+index, v.SpentTimeDeltaPercent)
		f.SetCellValue("All_stat", "K"+index, v.ChatCommentsTotal)
		f.SetCellValue("All_stat", "L"+index, v.ChatCommentsDeltaPercent)
		f.SetCellValue("All_stat", "M"+index, v.AnotherFields)
		f.SetCellValue("All_stat", "N"+index, v.BrowserClientInfo.UserIP)
		f.SetCellValue("All_stat", "O"+index, v.BrowserClientInfo.Platform)
		f.SetCellValue("All_stat", "P"+index, v.BrowserClientInfo.BrowserClient)
		f.SetCellValue("All_stat", "Q"+index, v.BrowserClientInfo.ScreenData_viewPort)
		f.SetCellValue("All_stat", "R"+index, v.BrowserClientInfo.ScreenData_resolution)
		//os loader
		f.SetCellValue("DB_stat", "A"+index, v.BrowserClientInfo.Platform)
		f.SetCellValue("DB_stat", "B"+index, v.BrowserClientInfo.BrowserClient)
		statPlatform[v.BrowserClientInfo.Platform]++
		statBrowser[v.BrowserClientInfo.BrowserClient]++
		//res loader
		f.SetCellValue("Resolution_stat", "A"+index, v.BrowserClientInfo.ScreenData_resolution)
		statResolution[v.BrowserClientInfo.ScreenData_resolution]++
		//ip loader
		if i < 1500 {
			ipip = append(ipip, v.BrowserClientInfo.UserIP)
		}
		//time loader
		if v.JoinTime[:5] != "0001-" && v.LeaveTime[:5] != "0001-" {
			timeStamps = append(timeStamps, stamp{v.JoinTime[:19], true, 0})
			timeStamps = append(timeStamps, stamp{v.LeaveTime[:19], false, 0})
		}
	}
	//os unloader
	//platform
	index := 2
	statOS := make(map[string]int)
	statOS["Android"] = 0
	statOS["iOS"] = 0
	statOS["OS X"] = 0
	statOS["Windows"] = 0
	statOS["Ubuntu"] = 0
	statOS["other"] = 0
	for i, v := range statPlatform {
		f.SetCellValue("DB_stat", "D"+strconv.Itoa(index), i)
		f.SetCellValue("DB_stat", "E"+strconv.Itoa(index), v)

		if i[:7] == "Android" {
			statOS["Android"] += v
		} else if i[:3] == "iOS" {
			statOS["iOS"] += v
		} else if i[:4] == "OS X" {
			statOS["OS X"] += v
		} else if i[:7] == "Windows" {
			statOS["Windows"] += v
		} else if i[:6] == "Ubuntu" {
			statOS["Ubuntu"] += v
		} else {
			statOS["other"] += v
		}

		index++
	}
	index = 2
	for i, v := range statOS {
		f.SetCellValue("DB_stat", "J"+strconv.Itoa(index), i)
		f.SetCellValue("DB_stat", "K"+strconv.Itoa(index), v)
		index++
	}
	//browser
	index = 2
	statBr := make(map[string]int)
	statBr["Microsoft Edge"] = 0
	statBr["Firefox"] = 0
	statBr["Chrome"] = 0
	statBr["Safari"] = 0
	statBr["Yandex Browser"] = 0
	statBr["Opera"] = 0
	statBr["other"] = 0
	for i, v := range statBrowser {
		f.SetCellValue("DB_stat", "G"+strconv.Itoa(index), i)
		f.SetCellValue("DB_stat", "H"+strconv.Itoa(index), v)

		if i[:9] == "Microsoft" {
			statBr["Microsoft Edge"] += v
		} else if i[:7] == "Firefox" {
			statBr["Firefox"] += v
		} else if i[:6] == "Chrome" {
			statBr["Chrome"] += v
		} else if i[:6] == "Safari" {
			statBr["Safari"] += v
		} else if i[:6] == "Yandex" {
			statBr["Yandex Browser"] += v
		} else if i[:5] == "Opera" {
			statBr["Opera"] += v
		} else {
			statBr["other"] += v
		}

		index++
	}
	index = 2
	for i, v := range statBr {
		f.SetCellValue("DB_stat", "M"+strconv.Itoa(index), i)
		f.SetCellValue("DB_stat", "N"+strconv.Itoa(index), v)
		index++
	}
	//res unloader
	index = 2
	statRatio := make(map[string]int)
	statRatio["16:9"] = 0
	statRatio["4:3"] = 0
	statRatio["1:1"] = 0
	statRatio["other"] = 0
	for i, v := range statResolution {
		f.SetCellValue("Resolution_stat", "C"+strconv.Itoa(index), i)
		f.SetCellValue("Resolution_stat", "D"+strconv.Itoa(index), v)
		x := strings.Index(i, "x")
		x1, _ := strconv.Atoi(i[:x])
		x2, _ := strconv.Atoi(i[x+1:])
		if x1*9 == x2*16 || x1*16 == x2*9 {
			statRatio["16:9"] += v
		} else if x1*4 == x2*3 || x1*3 == x2*4 {
			statRatio["4:3"] += v
		} else if x1 == x2 {
			statRatio["1:1"] += v
		} else {
			statRatio["other"] += v
		}

		index++
	}
	index = 2
	for i, v := range statRatio {
		f.SetCellValue("Resolution_stat", "F"+strconv.Itoa(index), i)
		f.SetCellValue("Resolution_stat", "G"+strconv.Itoa(index), v)
		index++
	}
	//ip unloader
	var ips []IP
	ALERT := 1
	index = 0
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
		f.SetCellValue("CP_stat", "A"+strconv.Itoa(i+2), v.Country)
		f.SetCellValue("CP_stat", "B"+strconv.Itoa(i+2), v.ISP)
		statLocation[v.Country]++
		statProvider[v.ISP]++
	}
	index = 2
	cRus := 0
	cOther := 0
	for i, v := range statLocation {
		f.SetCellValue("CP_stat", "D"+strconv.Itoa(index), i)
		f.SetCellValue("CP_stat", "E"+strconv.Itoa(index), v)
		if i == "Russia" {
			cRus += v
		} else {
			cOther += v
		}
		index++
	}
	f.SetCellValue("CP_stat", "J1", "Russia")
	f.SetCellValue("CP_stat", "J2", cRus)
	f.SetCellValue("CP_stat", "K1", "Other")
	f.SetCellValue("CP_stat", "K2", cOther)

	sCounty := `{"type": "col", "series": [
		{
			"name": "CP_stat!$J$1",
			"values": "CP_stat!$J$2"
			},
		{
			"name": "CP_stat!$K$1",
			"values": "CP_stat!$K$2"
		}
		],
		"title":
		{
			"name": "Location"
		}
		}`
	if err := f.AddChart("CP_stat", "L1", sCounty); err != nil {
		fmt.Println(err)
		return
	}

	index = 2
	for i, v := range statProvider {
		f.SetCellValue("CP_stat", "G"+strconv.Itoa(index), i)
		f.SetCellValue("CP_stat", "H"+strconv.Itoa(index), v)
		index++
	}
	//time unloader
	sort.Sort(timeStamps)
	counter := 0
	for i := range timeStamps {
		if timeStamps[i].start {
			counter++
			timeStamps[i].count = counter
		} else {
			counter--
			timeStamps[i].count = counter
		}
	}
	index = 2
	for i, v := range timeStamps[:timeStamps.Len()-1] {
		if v.time != timeStamps[i+1].time {
			f.SetCellValue("Peaks_stat", "A"+strconv.Itoa(index), v.time)
			f.SetCellValue("Peaks_stat", "B"+strconv.Itoa(index), timeStamps[i+1].time)
			f.SetCellValue("Peaks_stat", "C"+strconv.Itoa(index), v.count)
			index++
		}
	}
	sPeaks := `{"type": "line", "series": [
	{
		"name": "Peaks_stat!$A$1:$A$` + strconv.Itoa(index-1) + `",
		"values": "Peaks_stat!$C$1:$C$` + strconv.Itoa(index-1) + `"
		}],
		"title":
		{
			"name": "Peaks"
		}
		}`
	if err := f.AddChart("Peaks_stat", "E1", sPeaks); err != nil {
		fmt.Println(err)
		return
	}
	//save
	if err := f.SaveAs("All.xlsx"); err != nil {
		fmt.Println(err)
	}
	http.ServeFile(w, r, "All.xlsx")
}

type Port struct {
	Port string
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ping", pingHandler).Methods("GET")
	r.HandleFunc("/stat", statHandler).Methods("GET")
	r.HandleFunc("/collect", collectHandler).Methods("POST")
	r.HandleFunc("/report_all", reportAllHandler).Methods("GET")

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
