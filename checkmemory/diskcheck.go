package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rollbar/rollbar-go"
)

// DiskStatus datatype for disk space
type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// RollbarConnection initializes rollbar credentials, so we can be informed
// of all the errors this program encounters.
func RollbarConnection() {
	rollbar.SetToken(os.Getenv("ROLLBAR_CREDENTIALS"))
	rollbar.SetEnvironment(os.Getenv("ROLLBAR_ENV_TYPE"))    // defaults to "development"
	rollbar.SetCodeVersion("v1")                             // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerHost(os.Getenv("THOST"))                // optional override; defaults to hostname
	rollbar.SetServerRoot("github.com/dataf3l/monitor_tool") // path of project (required for GitHub integration and non-project stacktrace collapsing)
}

// DiskUsage of path/disk
func DiskUsage(path string) (DiskStatus, error) {
	var disk DiskStatus
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		// body := fmt.Sprintf("DiskUsage failed with an error: %w", err)
		return disk, err
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return disk, nil
}

// GetRequiredFreeDiskSpace return required free disk space from environment file
func GetRequiredFreeDiskSpace() (float64, error) {
	FreeDiskSpace := os.Getenv("FreeDiskSpace")
	if len(FreeDiskSpace) <= 0 {
		FreeDiskSpace = "2"
	}
	var FreeDiskSpaceFloat float64
	FreeDiskSpaceFloat, err := strconv.ParseFloat(FreeDiskSpace, 32)
	if err != nil {
		return FreeDiskSpaceFloat, err
	}
	return FreeDiskSpaceFloat, nil
}

// SendEmailNotification Sends an Email using HTTP
// This function should not be used, it is deprecated in favor of SendEmailNotification2
func SendEmailNotification(subject string, body string, to string) {
	from := os.Getenv("SMTP_FROM")
	pass := os.Getenv("SMTP_PASS")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"),
		smtp.PlainAuth("", os.Getenv("SMTP_USER"), pass, os.Getenv("SMTP_HOST")),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		rollbar.Critical(err)
		return
	}

	log.Print("email sent to " + to)
}

// CheckFreeDiskSpace check disk space for server and if its low terminate
func CheckFreeDiskSpace() {
	disk, err := DiskUsage("/")
	if err != nil {
		body := fmt.Sprintf("DiskUsage failed with an error: %w", err)
		rollbar.Critical(body)
		panic(err)
	}

	// Bits to GB
	freeSpace := float64(disk.Free) / float64(1073741824) // 1073741824 B = 1 GB
	usedSpace := float64(disk.Used) / float64(1073741824)
	totalDisk := float64(disk.All) / float64(1073741824)

	free, err := GetRequiredFreeDiskSpace()
	if err != nil {
		body := fmt.Sprintf("GetRequiredFreeDiskSpace failed with an error: %w", err)
		rollbar.Critical(body)
		panic(err)
	}

	if freeSpace > free {
		subject := "DISK SPACE FULL"
		// body := fmt.Sprintf("You have  used %f GB disk space and %f GB is free, please free some space for the server to continue ", usedSpace, freeSpace)
		body := fmt.Sprintf("You are currently using %f%% of your available memory", usedSpace/totalDisk*100)
		// fmt.Println("["+os.Getenv("THOST")+"]"+subject, body, os.Getenv("ADMIN_EMAIL"))
		SendEmailNotification("["+os.Getenv("THOST")+"]"+subject, body, os.Getenv("ADMIN_EMAIL"))
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// if the .env file does not exist, this program will fail.
		// administrators must ensure .env file exists in the same
		// location as the binary
		log.Println("null_value_nagger: critical: .env file is missing")
		panic(err)
	}
	RollbarConnection()
	CheckFreeDiskSpace()
	rollbar.Wait()
}
