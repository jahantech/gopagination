package main

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

/****  GZIP Http Response Writer ***********/
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}

/******** API Log File ***********/
func log_requests(ActiveUser string, log_data string) {
	currenttime := time.Now().Local()

	file, err := os.OpenFile("Web.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err = file.WriteString(currenttime.Format("15:04:05 01/02/2006") + "[" + ActiveUser + "]:" + log_data + "\r\n"); err != nil {
		panic(err)
	}
}

/***** Structs **********/
type PriceItemData struct {
	Data []ItemListVals `json:"data"`
}

type ItemListVals struct {
	SPC  string
	SPN  string
	SPPS string
	SPP  float64
}

/***** Main Function *****/
func main() {
	runtime.GOMAXPROCS(4)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/pricelist", makeGzipHandler(PriceList))
	http.HandleFunc("/pricelistquery", makeGzipHandler(PriceListQuery))

	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal(err)
	}
}

func PriceList(rw http.ResponseWriter, r *http.Request) {
	GetPage := r.URL.Query().Get("getpage")
	GetPageInt, _ := strconv.Atoi(GetPage)

	csvfile, err := os.Open("pricelist.csv")
	if err != nil {
		http.Error(rw, http.StatusText(405), 405)
		return
	}

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1
	fmt.Println("Reading data now")

	RawCSVdata, err := reader.ReadAll()

	if err != nil {
		http.Error(rw, err.Error(), 405)
		fmt.Println(err.Error())
		return
	}
	var oneRecord ItemListVals
	if GetPageInt == 0 {
		GetPageInt = 1
	}
	output_res_slice := make([]ItemListVals, 0)
	for jj, each := range RawCSVdata {

		if jj < (GetPageInt*50) && jj > ((GetPageInt-1)*50) {
			oneRecord.SPC = each[0]
			oneRecord.SPN = each[1]
			oneRecord.SPPS = each[2]
			oneRecord.SPP, _ = strconv.ParseFloat(each[3], 64)
			output_res_slice = append(output_res_slice, oneRecord)
		} else {
		}
	}
	allRecordsStruct := PriceItemData{output_res_slice}

	jsondata, err := json.Marshal(allRecordsStruct)

	if err != nil {
		http.Error(rw, http.StatusText(405), 405)
		return
	} else {
		fmt.Fprintf(rw, string(jsondata))
		return
	}

}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func PriceListQuery(rw http.ResponseWriter, r *http.Request) {
	GetQueryString := r.URL.Query().Get("querystr")
	var teststring []string
	GetPage := r.URL.Query().Get("getpage")
	GetPageInt, _ := strconv.Atoi(GetPage)
	csvfile, err := os.Open("pricelist.csv")
	if err != nil {
		http.Error(rw, http.StatusText(405), 405)
		return
	}
	if GetPageInt == 0 {
		GetPageInt = 1
	}
	defer csvfile.Close()
	iii := 0
	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1
	fmt.Println("Reading data now")
	RawCSVdata, err := reader.ReadAll()

	if err != nil {
		http.Error(rw, err.Error(), 405)
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Read:Success")
	var oneRecord ItemListVals

	teststring = append(teststring, GetQueryString)
	output_res_slice := make([]ItemListVals, 0)
	for _, each := range RawCSVdata {

		if strings.Contains(strings.ToLower(each[1]), strings.ToLower(GetQueryString)) {
			iii++
			if iii < (GetPageInt*50) && iii > ((GetPageInt-1)*50) {
				oneRecord.SPC = each[0]
				oneRecord.SPN = each[1]
				oneRecord.SPPS = each[2]
				oneRecord.SPP, _ = strconv.ParseFloat(each[3], 64)
				output_res_slice = append(output_res_slice, oneRecord)
			}
		} else {
		}
	}
	allRecordsStruct := PriceItemData{output_res_slice}

	jsondata, err := json.Marshal(allRecordsStruct)
	fmt.Println("Completed")
	fmt.Println(iii)

	if err != nil {
		http.Error(rw, http.StatusText(405), 405)
		return
	} else {
		fmt.Fprintf(rw, string(jsondata))
		return
	}

}
