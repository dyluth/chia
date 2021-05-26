package main

import (
	"bytes"
	"fmt"
	exec "os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	failedRecoveries = 0
	allRecoveries    = 0
)

func main() {
	ticker := time.NewTicker(2 * time.Minute)
	CheckAndRecover()
	for {
		select {
		case <-ticker.C:
			CheckAndRecover()
		}
	}

}

func CheckAndRecover() {
	out, err := RunCmd("chia", "farm", "summary")
	//fmt.Printf("ran.. %v, %v\n", out, err)

	statusRe := regexp.MustCompile(`Farming status:\s+(.+)\n`)
	statusMatch := statusRe.FindStringSubmatch(out)
	if len(statusMatch) < 2 {
		fmt.Printf("======\nEEK! dont know what to do with output \n`%v`\n we matched: [%v]\n======\n", out, strings.Join(statusMatch, "]["))
		return
	}
	statusString := statusMatch[1]

	if err != nil {
		fmt.Printf("broke?? %+v\n", err)
		return
	}

	if strings.Contains(statusString, "Farming") {
		fmt.Printf("looks good to me :)\n")
	} else if strings.Contains(statusString, "Not synced or not connected to peers") {
		fmt.Printf("NOT SYNCED!! trying to recover!")
		allRecoveries++
		err = RecoverChia()
		if err != nil {
			fmt.Printf("Failed To recover: %v\n", err)
			failedRecoveries++
		}
		fmt.Printf("TotalRecoveries: %v, Failed: %v\n", allRecoveries, failedRecoveries)
	} else {
		fmt.Printf("not going to take action on status: `%v`\n", statusString)
	}
}

func RunCmd(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func RecoverChia() error {
	out, err := RunCmd("chia", "show", "-c")
	fmt.Printf("RECOVERING\n")
	if err != nil {
		fmt.Printf("broke?? %+v\n", err)
		return err
	}
	// count instances of `FULL_NODE node.chia.net`
	nodeRE := regexp.MustCompile("FULL_NODE node.chia.net")
	matches := nodeRE.FindAllString(out, -1)
	if len(matches) < 2 {
		fmt.Printf("only %v good nodes to sync from.. adding\n", len(matches))
		out, err = RunCmd("chia", "show", "-a", "node.chia.net:8444")
		if err != nil {
			fmt.Printf("failed to add node somehow %v\n", err)
			return err
		}
		fmt.Printf("added node.chia.net:8444: `%v` hope that helps...\n", out)
	}
	return nil
}
