package main

import (
	"flag"
	"fmt"
	//"log"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	//"path/filepath"
	//"sort"
	"net/http"
	//"strings"
	"text/tabwriter"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type OctoprintStatus struct {
	State struct {
		Error string `json:"error"`
		Flags struct {
			Cancelling    bool `json:"cancelling"`
			ClosedOrError bool `json:"closedOrError"`
			Error         bool `json:"error"`
			Finishing     bool `json:"finishing"`
			Operational   bool `json:"operational"`
			Paused        bool `json:"paused"`
			Pausing       bool `json:"pausing"`
			Printing      bool `json:"printing"`
			Ready         bool `json:"ready"`
			Resuming      bool `json:"resuming"`
			SdReady       bool `json:"sdReady"`
		} `json:"flags"`
		Text string `json:"text"`
	} `json:"state"`
}

type GettererPrinterList struct {
	Printers []struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	} `json:"printers"`
}

func init() {
	flag.String("api", "statustoken", "Getterer API token")
	flag.String("gettererurl", "http://172.28.0.10:54038", "Getterer URL")
	//flag.String("gettererurl", "http://127.0.0.1:54038", "Getterer URL")
	flag.Int("ttl", 10, "TTL")
	flag.Int("padding", 2, "Column padding")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {

	//tempPrinterList := getURL("http://127.0.0.1:54038/printers?json=y")
	tempPrinterList := getURL("http://172.28.0.10:54038/printers?json=y")
	//fmt.Println(prettyPrint(tempPrinterList))

	/*
		for _, line := range strings.Split(strings.TrimSuffix(tempPrinterList, "\n"), "\n") {
			fmt.Printf("--%s|    %s\n", line, strings.TrimPrefix(line, ":"))
		}
	*/

	allPrinters := GettererPrinterList{}
	err := json.Unmarshal([]byte(tempPrinterList), &allPrinters)

	if err != nil {
		log.Fatal(err)
	}

	/*
		fmt.Printf("====\n%s\n", prettyPrint(allPrinters))

		fmt.Printf("=============\n%#v\n=========\n", allPrinters)
		//fmt.Printf("ramonname=%s\ndesc=%s\n", allPrinters.Printers["ramon"].Name, allPrinters.Printers["ramon"].Desc)

		spew.Dump(allPrinters.Printers)

		fmt.Printf("length of allPrinters.Printers=%d\n", len(allPrinters.Printers))

	*/

	w := tabwriter.NewWriter(os.Stdout, 0, 4, viper.GetInt("padding"), ' ', 0)
	columnheadings := "Printer\tDesc\tStatus\tOther Status\tError\n"
	fmt.Fprint(w, columnheadings)

	url := viper.GetString("gettererurl") + "/status/"

	for _, something := range allPrinters.Printers {

		//fmt.Printf("Name: %s      Desc: %s\n", something.Name, something.Desc)

		result := getURL(url + something.Name + "?json=y")

		name := something.Name
		desc := something.Desc
		state := getURL(url + something.Name)

		currentPrinterState := OctoprintStatus{}
		err = json.Unmarshal([]byte(result), &currentPrinterState)

		fmt.Println(prettyPrint(currentPrinterState))

		if err != nil {
			log.Fatal(err)
		}

		stateOther := currentPrinterState.State.Text
		printerError := currentPrinterState.State.Error

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, desc, state, stateOther, printerError)

	}
	//os.Exit(0)

	//==============

	/*
		fmt.Printf("ttl: %d\n", viper.GetInt("ttl"))
		url := viper.GetString("gettererurl") + "/status/"

		fmt.Println("Ramon:")
		fmt.Printf("ramon getsingle=%s\n", getURL(url+"ramon"))

		fmt.Println("\n\nEnder5:")
		fmt.Printf("ender5 getsingle=%s\n", getURL(url+"ender5"))

		fmt.Println("\n\nSandy (single):")
		fmt.Printf("sandy getsingle=%s\n", getURL(url+"sandy"))

		fmt.Println("\n\nSandy (json):")
		result := getURL(url + "sandy?json=y")
		fmt.Printf("sandy getsingle=%s\n", result)

		currentPrinterState := OctoprintStatus{}
		err = json.Unmarshal([]byte(result), &currentPrinterState)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Printer: %s\n", currentPrinterState.State.Text)

		if currentPrinterState.State.Error == "" {
			fmt.Printf("  Error: %s\n", currentPrinterState.State.Error)
		}
	*/

	w.Flush()

}

func getURL(url string) string {
	client := http.Client{
		Timeout: time.Duration(viper.GetInt("ttl")) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error: Could not create client connection to %s\n", url)
		return "error"
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Printf("Error: Cannot take getSingle URL \"%s\", Message:%s\n", url, err)
			return "Could not take snapshot"
		}

		return string(body)
	} else {
		fmt.Println("Error: Could not getSingle: " + url + " HTTPStatus: " + string(resp.StatusCode))
		return "Could not getSingle"
	}
}

// prints out structs or json prettily
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
