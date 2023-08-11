package main


import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/json"
	"io/ioutil"

	"github.com/shirou/gopsutil/process"
	
)

type ProcessInfo struct {
	PID       int32         `json:"pid"`
	Exe       string        `json:"exe"`
	Active    bool          `json:"active"`
	TotalTime time.Duration `json:"total_time"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	processes := make(map[int32]struct {
		StartTime time.Time
		IsActive  bool
		TotalTime time.Duration
	})

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			psList, err := process.Processes()
			if err != nil {
				log.Println("Error getting processes:", err)
				continue
			}

			var processInfoList []ProcessInfo

			for _, ps := range psList {
				pid := ps.Pid
				state, err := ps.Status()
				if err != nil {
					log.Println("Error")
				}

				exe, exeErr := ps.Exe()
				if exeErr != nil {
					log.Println("Error")
				}

				info, found := processes[pid]
				if !found {
					info = struct {
						StartTime time.Time
						IsActive  bool
						TotalTime time.Duration
					}{StartTime: time.Now(), IsActive: false, TotalTime: 0}
				}

				if state == "R" || state == "S" {
					info.IsActive = true
				} else {
					info.IsActive = false
				}

				if info.IsActive {
					info.TotalTime += time.Since(info.StartTime)
				}

				processes[pid] = info

				processInfoList = append(processInfoList, ProcessInfo{
					PID:       pid,
					Exe:       exe,
					Active:    info.IsActive,
					TotalTime: info.TotalTime,
				})
			}

			fmt.Println("Process Info:")
			for _, info := range processInfoList {
				fmt.Printf("PID: %d, Exe: %s, Active: %t, Total Time: %s\n", info.PID, info.Exe, info.Active, info.TotalTime)
			}

			

			err = saveProcessInfoToJSON(processInfoList)
			if err != nil {
				log.Println("Error saving process info to JSON:", err)
			}

		case <-interrupt:
			fmt.Println("Exiting...")
			return
		}
	}
}

func saveProcessInfoToJSON(processInfoList []ProcessInfo) error {
	jsonData, err := json.Marshal(processInfoList)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("process_info.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

