package podwiz

import (
	"fmt"
	"github.com/Atish03/podwiz/reqProto"
	"net"
	"io"
	"google.golang.org/protobuf/proto"
	"path/filepath"
	"encoding/json"
	"github.com/olekukonko/tablewriter"
	"os"
)

type Creds struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Port uint16 `json:"ports"`
}

type ScheduleInfo struct {
	StartTime string `json:startTime`
	EndTime string `json:endTime`
	Name string `json:name`
	PodName string `json:podName`
}

type Received struct {
	Command string `json:command`
	Data []byte `json:data`
}

type Socket struct {
	Socket *net.Conn
}

func reader(r io.Reader) Received {
    buf := make([]byte, 4096)
	data := Received{}
	n, err := r.Read(buf[:])
	if err != nil {
		return
	}
	err = json.Unmarshal(buf[0:n], &data)
	if err != nil {
		fmt.Println(string(buf[0:n]))
	}

	if data.Command == "start" {
		creds := Creds{}
		err = json.Unmarshal(data.Data, &creds)
		if err != nil {
			fmt.Println("Server didnt send correct data!")
		}
		fmt.Printf("Username: %s\nPassword: %s\nPort: %d\n", creds.Username, creds.Password, creds.Port)
		return creds
	} else if data.Command == "list" {
		schedules := []ScheduleInfo{}
		err = json.Unmarshal(data.Data, &schedules)
		if err != nil {
			fmt.Println("Server didnt send correct data!")
		}
		toShow := [][]string{}
		for i := 0; i < len(schedules); i++ {
			toShow = append(toShow, []string{
				schedules[i].Name,
				schedules[i].PodName,
				schedules[i].StartTime,
				schedules[i].EndTime,
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Pod name", "Start", "Finish"})

		for _, v := range toShow {
			table.Append(v)
		}
		table.Render() 
	}
}

func Connect() *Socket {
	socket, err := net.Dial("unix", "/tmp/podwiz.sock")
	if err != nil {
		fmt.Println("Cannot connect to podwiz!\nAre you sure it is running?")
		return nil
	}

	sock := Socket {
		&socket,
	}

	return &sock
}

func (socket *Socket) send(out []byte) Received {
	for {
        _, err := (*socket.Socket).Write(out)
        if err != nil {
            break
        }
    }

	reader(*(socket.Socket))

	return []byte("none")
}

func Start(name string, machineName string, path string, imgName string, time int, scheduleName string) []byte {
	absPath, err := filepath.Abs(path)
    if err != nil {
        panic(err)
    }

	block := &reqProto.Block{
		Command: "start",
		Start: &reqProto.Start {
			Name: name,
			MachineName: machineName,
			Path: absPath,
			ImgName: imgName,
			Time: int64(time),
			ScheduleName: scheduleName,
		},
	}

	out, err := proto.Marshal(block)
	if err != nil {
		panic(err)
	}

	data := send(out)

	return data.Data
}

func List(scheduleName string) []byte {
	block := &reqProto.Block {
		Command: "list",
		List: &reqProto.List {
			ScheduleName: scheduleName,
		},
	}

	out, err := proto.Marshal(block)
	if err != nil {
		panic(err)
	}

	data := send(out)

	return data.Data
}