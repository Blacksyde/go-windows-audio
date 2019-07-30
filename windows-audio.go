//Author: Alex Blacketor

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
	"github.com/eiannone/keyboard"
)

type WindowsAudioDevice struct {
	Index   int    `json:"index"`
	Default bool   `json:"default"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	ID      string `json:"id"`
	Device  string `json:"device"`
}

var (
	shell ps.Shell
	back  *backend.Local

	devices       []WindowsAudioDevice
	currentDevice WindowsAudioDevice

	err error
)

func initPowerShell() {
	shell, err = ps.New(back)
	if err != nil {
		panic(err)
	}

	// Commands to import the audio device module
	_, _, err = shell.Execute("New-Item \"$($profile | split-path)\\Modules\\AudioDeviceCmdlets\" -Type directory -Force")
	if err != nil {
		panic(err)
	}
	_, _, err = shell.Execute("Copy-Item \"AudioDeviceCmdlets.dll\" \"$($profile | split-path)\\Modules\\AudioDeviceCmdlets\\AudioDeviceCmdlets.dll\"")
	if err != nil {
		panic(err)
	}
	_, _, err = shell.Execute("Set-Location \"$($profile | Split-Path)\\Modules\\AudioDeviceCmdlets\"")
	if err != nil {
		panic(err)
	}
	_, _, err = shell.Execute("Get-ChildItem | Unblock-File")
	if err != nil {
		panic(err)
	}
	_, _, err = shell.Execute("Import-Module AudioDeviceCmdlets")
	if err != nil {
		panic(err)
	}
}

func populateDevices() {
	stdout, _, err := shell.Execute("Get-AudioDevice -List")
	if err != nil {
		panic(err)
	}

	a := regexp.MustCompile(`\r\n\r\n`)
	//fmt.Printf("%q\n", a.Split(stdout, -1))

	audioDevices := a.Split(stdout, -1)
	audioDevices = audioDevices[1 : len(audioDevices)-1]
	//fmt.Println(audioDevices[0])

	b := regexp.MustCompile(`\r\n`)
	c := regexp.MustCompile(`[\s]*:[\s]*`)

	var deviceList [][][]string
	deviceList = make([][][]string, len(audioDevices))

	for i, device := range audioDevices {

		deviceLines := b.Split(device, -1)
		//fmt.Println(i, deviceLines)

		deviceList[i] = make([][]string, len(deviceLines))

		for j, line := range deviceLines {

			deviceFields := c.Split(line, 2)
			//fmt.Println("	", j, deviceFields)
			deviceList[i][j] = make([]string, len(deviceFields))
			deviceList[i][j] = deviceFields
		}
	}

	//fmt.Println("Field 0 of device at index 0:", deviceList[0][0])

	devices = make([]WindowsAudioDevice, len(deviceList))

	for i, d := range deviceList {
		var device WindowsAudioDevice
		device.Index, err = strconv.Atoi(d[0][1])
		if err != nil {
			panic(err)
		}
		device.Default, err = strconv.ParseBool(d[1][1])
		if err != nil {
			panic(err)
		}
		device.Type = d[2][1]
		device.Name = d[3][1]
		device.ID = d[4][1]
		device.Device = d[5][1]

		devices[i] = device
	}

	//fmt.Println(devices)
}

func init() {
	back = &backend.Local{}
	initPowerShell()
	populateDevices()
}

func main() {
	//Exit powershell when the program exits
	defer shell.Exit()

	//printAudioDeviceList()

	//setAudioDeviceByIndex(devices[2].Index)
	//setAudioDeviceByIndex(devices[3].Index)

	//printPlaybackDevice()
	//printRecordingDevice()

	//setAudioDeviceByID(devices[0].ID)

	listenForKeys()
}

func setAudioDeviceByIndex(index int) {
	_, _, err := shell.Execute("Set-AudioDevice -Index " + strconv.Itoa(index))
	if err != nil {
		panic(err)
	}
	fmt.Println(devices[index-1].Type, "device set to", devices[index-1].Name, "\n")
}

func setAudioDeviceByID(ID string) {
	_, _, err := shell.Execute("Set-AudioDevice -ID " + "\"" + ID + "\"")
	if err != nil {
		panic(err)
	}
	fmt.Println("Audio device with ID", ID, "enabled", "\n")
}

func printAudioDeviceList() {
	stdout, _, err := shell.Execute("Get-AudioDevice -List")
	if err != nil {
		panic(err)
	}

	fmt.Println(stdout)
}

func printPlaybackDevice() {
	stdout, _, err := shell.Execute("Get-AudioDevice -Playback")
	if err != nil {
		panic(err)
	}

	fmt.Println(stdout)
}

func playbackMute() bool {
	stdout, _, err := shell.Execute("Get-AudioDevice -PlaybackMute")
	if err != nil {
		panic(err)
	}
	out := strings.TrimSpace(stdout)
	b, err := strconv.ParseBool(out)
	return b
}

func playbackMuteToggle() {
	_, _, err := shell.Execute("Set-AudioDevice -PlaybackMuteToggle")
	if err != nil {
		panic(err)
	}
	fmt.Println("Playback device mute toggled.\n")
}

func printRecordingDevice() {
	stdout, _, err := shell.Execute("Get-AudioDevice -Recording")
	if err != nil {
		panic(err)
	}

	fmt.Println(stdout)
}

func recordingMute() bool {
	stdout, _, err := shell.Execute("Get-AudioDevice -RecordingMute")
	if err != nil {
		panic(err)
	}
	out := strings.TrimSpace(stdout)
	b, err := strconv.ParseBool(out)
	if err != nil {
		panic(err)
	}
	return b
}

func listenForKeys() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	fmt.Println("Press ESC to quit\n")
	for {
		char, key, err := keyboard.GetKey()
		//fmt.Printf("You pressed: %q\r\n", char)
		if err != nil {
			panic(err)
		} else if key == keyboard.KeyEsc {
			break
		} else {
			ind := int(char - '0')
			//fmt.Println(ind)
			if ind >= 1 && ind < len(devices)+1 {
				setAudioDeviceByIndex(ind)
			} else {
				st := string(char)
				if st == "p" {
					fmt.Println("Current playback device:")
					printPlaybackDevice()
					fmt.Println("Muted:", playbackMute(), "\n")
				}
				if st == "r" {
					fmt.Println("Current recording device:")
					printRecordingDevice()
					fmt.Println("Muted:", recordingMute(), "\n")
				}
				if st == "l" {
					fmt.Println("Audio device list:")
					printAudioDeviceList()
				}
				if st == "m" {
					playbackMuteToggle()
				}
			}
		}
	}
}

/*
Get-AudioDevice   -List             		# Outputs a list of all devices as <AudioDevice>
                  -ID <string>      		# Outputs the device with the ID corresponding to the given <string>
                  -Index <int>      		# Outputs the device with the Index corresponding to the given <int>
				  -Playback         		# Outputs the default playback device as <AudioDevice>
                  -PlaybackMute     		# Outputs the default playback device's mute state as <bool>
                  -PlaybackVolume   		# Outputs the default playback device's volume level on 100 as <float>
                  -Recording        		# Outputs the default recording device as <AudioDevice>
                  -RecordingMute    		# Outputs the default recording device's mute state as <bool>
				  -RecordingVolume  		# Outputs the default recording device's volume level on 100 as <float>

Set-AudioDevice   <AudioDevice>             # Sets the default playback/recording device to the given <AudioDevice>, can be piped
                  -ID <string>              # Sets the default playback/recording device to the device with the ID corresponding to the given <string>
                  -Index <int>              # Sets the default playback/recording device to the device with the Index corresponding to the given <int>
                  -PlaybackMute <bool>      # Sets the default playback device's mute state to the given <bool>
                  -PlaybackMuteToggle       # Toggles the default playback device's mute state
                  -PlaybackVolume <float>   # Sets the default playback device's volume level on 100 to the given <float>
                  -RecordingMute <bool>     # Sets the default recording device's mute state to the given <bool>
                  -RecordingMuteToggle      # Toggles the default recording device's mute state
				  -RecordingVolume <float>  # Sets the default recording device's volume level on 100 to the given <float>

Write-AudioDevice -PlaybackMeter  			# Writes the default playback device's power output on 100 as a meter
                  -PlaybackSteam  			# Writes the default playback device's power output on 100 as a stream of <int>
                  -RecordingMeter 			# Writes the default recording device's power output on 100 as a meter
				  -RecordingSteam 			# Writes the default recording device's power output on 100 as a stream of <int>
*/
