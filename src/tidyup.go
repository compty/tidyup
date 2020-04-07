package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

type DirectoryEntry struct {
	WatchPath       string
	DestinationPath string
	hours           int
}

type FileEntry struct {
	Path        string
	TimeChecked string
	Ignore      bool
}

func (de *DirectoryEntry) serialize() string {
	return de.WatchPath + "," + de.DestinationPath + "," + string(de.hours)
}

func deserializeDirectoryEntry(input string) (bool, DirectoryEntry) {
	entry := strings.SplitN(input, ",", 3)

	if len(entry) != 3 {
		return true, DirectoryEntry{}
	}

	// TODO: handle error if hours field is not an int
	hours, _ := strconv.Atoi(entry[2])

	newde := DirectoryEntry{entry[0], entry[1], hours}

	return false, newde
}

func (fe *FileEntry) serialize() string {
	return fe.Path + "," + fe.TimeChecked + "," + strconv.FormatBool(fe.Ignore)
}

func deserializeFileEntry(input string) FileEntry {
	entry := strings.SplitN(input, ",", 3)

	// TODO: handle error if ignore field is not a bool
	ignore, _ := strconv.ParseBool(entry[2])

	newfe := FileEntry{entry[0], entry[1], ignore}

	return newfe
}

func main() {
	fmt.Println("Tidy up")

	onExit := func() {
		fmt.Println("Starting onExit")
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
		fmt.Println("Finished onExit")
	}
	// Should be called at the very beginning of main().
	systray.RunWithAppWindow("TidyUp", 1024, 768, onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(Data, Data)
	mStartup := systray.AddMenuItem("Enable on startup", "Enable on startup")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mStartup.ClickedCh:
				fmt.Println("Enable on startup clicked")
			case <-mQuit.ClickedCh:
				systray.Quit()
				fmt.Println("Quit now...")
				return
			}
		}
	}()

	fileEntryStateFile := "/Users/Dave/Library/Application Support/TidyUp/tidyupfilelist"
	directoryEntryStateFile := "/Users/Dave/Library/Application Support/TidyUp/tidyupwatchlist"

	fmt.Println("Loading state")

	state, filelist, watchList := loadState(fileEntryStateFile, directoryEntryStateFile)

	if !state {
		emptyFileList := []FileEntry{}
		emptyWatchList := []DirectoryEntry{}
		saveState(fileEntryStateFile, directoryEntryStateFile, emptyFileList, emptyWatchList)
	}

	for _, f := range filelist {
		fmt.Println(f)
	}

	for _, f := range watchList {
		fmt.Println("Here")
		fmt.Println(f)
	}
}

func saveToFile(filename string, entries []string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	for _, v := range entries {
		fmt.Fprintln(f, v)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func saveState(fileEntryStateFile string, directoryEntryStateFile string, fileList []FileEntry, watchList []DirectoryEntry) {

	fileEntryAsStrings := []string{}

	for _, entry := range fileList {
		fileEntryAsStrings = append(fileEntryAsStrings, entry.serialize())
	}

	saveToFile(fileEntryStateFile, fileEntryAsStrings)

	watchListAsStrings := []string{}

	for _, entry := range watchList {
		watchListAsStrings = append(watchListAsStrings, entry.serialize())
	}

	fmt.Println("all files written successfully")
}

func loadFile(filename string) []string {

	fmt.Println("Loading file: ", filename)

	toReturn := []string{}

	file, err := os.Open(filename)
	if err != nil {
		return toReturn
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		// ignore lines with a # as these are comments
		if !strings.HasPrefix(scanner.Text(), "#") {
			if len(scanner.Text()) > 0 {
				toReturn = append(toReturn, scanner.Text())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return toReturn
}

func loadState(fileEntryStateFile string, directoryEntryStateFile string) (bool, []FileEntry, []DirectoryEntry) {

	toReturnFileEntries := []FileEntry{}
	toReturnDirectoryEntires := []DirectoryEntry{}

	fileEntriesAsString := loadFile(fileEntryStateFile)

	for _, fileEntry := range fileEntriesAsString {
		toReturnFileEntries = append(toReturnFileEntries, deserializeFileEntry(fileEntry))
	}

	directoryEntriesAsString := loadFile(directoryEntryStateFile)

	for _, directoryEntry := range directoryEntriesAsString {
		err, dirEntry := deserializeDirectoryEntry(directoryEntry)
		if err {
			// TODO: Probably a new line. Nothing really to do?
		}
		toReturnDirectoryEntires = append(toReturnDirectoryEntires, dirEntry)
	}

	return true, toReturnFileEntries, toReturnDirectoryEntires
}
