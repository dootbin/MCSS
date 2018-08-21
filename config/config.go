package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	//ServerRootDirectory contains folder with server folder
	ServerRootDirectory string
	//ServerName name of directory spigot.jar is located - name of your server
	ServerName string
	//SaveDir contains output save directory
	SaveDir string
	//SaveDiameter Diamater of target save
	SaveDiameter string
	//SaveName Name that will prefix save files
	SaveName string
	//FTPUser Ftp username
	FTPUser string
	//FTPPassword Ftp password
	FTPPassword string
	//FTPURL FTP url
	FTPURL string
	//WorldName world name
	WorldName string
	//WebHookURL webhook url string
	WebHookURL string

	config *configStruct
)

type configStruct struct {
	SaveDir             string `json:"SaveDir"`
	ServerRootDirectory string `json:"ServerRootDirectory"`
	ServerName          string `json:"ServerName"`
	SaveDiameter        string `json:"SaveDiameter"`
	SaveName            string `json:"SaveName"`
	FTPPassword         string `json:"FTPPassword"`
	FTPUser             string `json:"FTPUser"`
	FTPURL              string `json:"FTPURL"`
	WorldName           string `json:"WorldName"`
	WebHookURL          string `json:"WebHookURL"`
}

//ReadConfig reads config.json file.
func ReadConfig() error {

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	ServerRootDirectory = config.ServerRootDirectory
	SaveDir = config.SaveDir
	ServerName = config.ServerName
	SaveDiameter = config.SaveDiameter
	SaveName = config.SaveName
	FTPPassword = config.FTPPassword
	FTPURL = config.FTPURL
	WorldName = config.WorldName
	FTPUser = config.FTPUser
	WebHookURL = config.WebHookURL

	return nil
}
