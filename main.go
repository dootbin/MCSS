/*

MCRS - Minecraft Save Shit
Author: Benjamin Miles
Date: 8.1.2018

Notes: Program to backup select region files and player data to archive.


*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"./config"
	"./copy"
	"./messenger"
	"github.com/jinzhu/now"
	"github.com/mholt/archiver"
)

//LogReport Report that will be sent out.
var LogReport string

// Copy region files
func copyRegions(tmp string, targetDirectory string, saveDiameter int) {

	//16blocks in a chunk 32 chunks in a region bitshift.
	regions := saveDiameter >> 9
	counter := 0
	regionVerifyTarget := (regions + (regions + 2)) * (regions + (regions + 2))

	for i := ((regions + 1) * -1); i <= regions; i++ {

		for a := ((regions + 1) * -1); a <= regions; a++ {

			filename := "r." + strconv.Itoa(i) + "." + strconv.Itoa(a) + ".mca"

			targetFile := targetDirectory + "/" + filename
			saveTarget := tmp + "/" + filename

			err := copy.Copy(targetFile, saveTarget)

			if err != nil {

				LogReport += filename + " FAILED TO COPY\n"

			} else {
				counter++
			}
		}
	}

	LogReport += "Number of regions = " + strconv.Itoa(counter) + "\n"
	LogReport += "Target of regions = " + strconv.Itoa(regionVerifyTarget) + "\n"

	if counter == regionVerifyTarget {
		LogReport += "You have the correct number of region files\n"
	} else {

		LogReport += "You have missed the target number of regions\n"
	}

}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//ftpDelete crafts and execute curl command to delete file off host ftp
func ftpDelete(userName string, password string, host string, filename string) string {
	/* this feels like a hack */
	path, _ := exec.LookPath("curl")

	hostURL := fmt.Sprintf("ftp://%s/%s", host, filename)
	usernamePassword := fmt.Sprintf("%s:%s", userName, password)
	ftpDeleteFile := fmt.Sprintf("DELE /%s", filename)

	out, err := exec.Command(path, "-v", hostURL, "--user", usernamePassword, "-O", "--quote", ftpDeleteFile).Output()

	if err != nil {
		return err.Error()
	}
	return string(out)
}

//ftpSave crafts and execute curl command for xfer of file to host ftp
func ftpSave(userName string, password string, host string, source string, filename string) string {

	path, _ := exec.LookPath("curl")
	commandString := fmt.Sprintf("-T %s ftp://%s/%s --user %s:%s", source, host, filename, userName, password)
	command := strings.Split(commandString, " ")
	out, err := exec.Command(path, command[0:]...).Output()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func main() {

	//Read Config
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
	}

	saveDiameter, err := strconv.Atoi(config.SaveDiameter)

	//var saveDirectory = config.SaveDir
	var serverRootDirectory = config.ServerRootDirectory
	var serverName = config.ServerName
	var saveName = config.SaveName
	var ftpUser = config.FTPUser
	var ftpPassword = config.FTPPassword
	var ftpURL = config.FTPURL
	var worldName = config.WorldName
	var webHookURL = config.WebHookURL
	currentTime := time.Now()
	d := currentTime.Day()
	m := int(currentTime.Month())
	y := currentTime.Year()
	//End of Month
	eom := int(now.EndOfMonth().Day())
	lastMonth := m - 1
	//temperary working directories
	tmp := fmt.Sprintf("/%s/tmp", serverRootDirectory)
	save := fmt.Sprintf("/%s/save", serverRootDirectory)
	isSaveFolderVerified, _ := exists(save)
	isTmpFolderVerified, _ := exists(tmp)

	//Verify that there is a clean directory to work in.
	if isTmpFolderVerified {

		os.RemoveAll(tmp)
		os.MkdirAll(tmp, os.ModePerm)

	} else {

		os.MkdirAll(tmp, os.ModePerm)

	}

	if isSaveFolderVerified {

		os.RemoveAll(save)
		os.MkdirAll(save, os.ModePerm)

	} else {

		os.MkdirAll(save, os.ModePerm)

	}

	//Save Nether = SaveDiameter/8
	netherDiameter := saveDiameter / 8
	netherSource := fmt.Sprintf("/%s/%s/%s_nether/region", serverRootDirectory, serverName, worldName)

	//copy nether regions to tmp folder.
	netherDest := fmt.Sprintf("/%s/%s_nether/region", tmp, worldName)
	LogReport += "Attempted to Copy Nether\n"
	copyRegions(netherDest, netherSource, netherDiameter)

	//copy overworld regions to tmp folder.
	overworldSource := fmt.Sprintf("/%s/%s/%s/region", serverRootDirectory, serverName, worldName)
	overworldDest := fmt.Sprintf("/%s/%s/region", tmp, worldName)
	LogReport += "Attempted to Copy OverWorld\n"
	copyRegions(overworldDest, overworldSource, saveDiameter)

	//Copy playerdata over to tmp folder
	playerDataLocation := fmt.Sprintf("/%s/%s/%s/playerdata", serverRootDirectory, serverName, worldName)
	playerDataTmpLocation := fmt.Sprintf("/%s/%s/playerdata", tmp, worldName)
	LogReport += "Attempted to Copy PlayerData\n"
	copy.Copy(playerDataLocation, playerDataTmpLocation)

	//Create backup name
	mString := strconv.Itoa(m)
	dString := strconv.Itoa(d)
	yString := strconv.Itoa(y)
	backupName := fmt.Sprintf("%s.%s.%s.%s.tar.gz", saveName, mString, dString, yString)

	//Compress backup
	backupDest := fmt.Sprintf("%s/%s", save, backupName)
	err = archiver.TarGz.Make(backupDest, []string{tmp})

	// Send backup to ftp server
	a := ftpSave(ftpUser, ftpPassword, ftpURL, backupDest, backupName)
	if a == "" {

		//Send success message
		LogReport += "Save and FTP Sucess\n"
	} else {

		LogReport += "Save and FTP FAILED\n"
	}

	//Delete Old Backup. Keep one month's worth of backups and additionally a years worth of monthly backups (12)
	if lastMonth == 0 {
		lastMonth = 12
	}

	var fileToDelete string
	if d == eom {

		deleteCounter := 31 - d
		for i := 0; i <= deleteCounter; i++ {

			fileToDelete = saveName + strconv.Itoa(lastMonth) + "." + strconv.Itoa(d+i) + "." + strconv.Itoa(y) + ".tar.gz"
			f := ftpDelete(ftpUser, ftpPassword, ftpURL, fileToDelete)
			println(f)
		}

	} else {

		if d < 1 {

			fileToDelete = saveName + strconv.Itoa(lastMonth) + "." + strconv.Itoa(d) + "." + strconv.Itoa(y) + ".tar.gz"
			f := ftpDelete(ftpUser, ftpPassword, ftpURL, fileToDelete)
			println(f)
		} else {

			fileToDelete = saveName + strconv.Itoa(m) + "." + strconv.Itoa(d) + "." + strconv.Itoa(y-1) + ".tar.gz"
			f := ftpDelete(ftpUser, ftpPassword, ftpURL, fileToDelete)
			println(f)
		}

	}

	LogReport += "Finished Save"
	messenger.DiscordMessage(LogReport, webHookURL)

}
