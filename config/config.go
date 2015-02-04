package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/fcoury/rtc-go/rtc"
)

type Config struct {
	User    string `json:"user"`
	Pass    string `json:"pass"`
	OwnerId string `json:"rtcOwnerId"`
}

func ReadConfig() (*Config, error) {
	var c *Config

	file, err := configFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		err = CreateConfig()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func read(s string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s ", s)
	r, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	r = strings.Trim(r, "\n\r")
	if err != nil {
		return "", err
	}

	return r, nil
}

func CreateConfig() error {
	file, err := configFile()
	if err != nil {
		return err
	}

	user, err := read("RTC username:")
	if err != nil {
		return err
	}
	pass, err := read("RTC password:")
	if err != nil {
		return err
	}

	r := rtc.NewRTC(user, pass)
	err = r.Login()
	if err != nil {
		return err
	}

	data, err := r.GetAllValues()
	if err != nil {
		return err
	}

	// fmt.Printf("%+v\n", data)

	fmt.Println("\nPlease select the user you want to be the owner of your work items:")
	var owners []string
	i := 1
	for k, v := range data["owner"] {
		owners = append(owners, k)
		fmt.Printf("  %d. %s\n", i, v)
		i = i + 1
	}

	ownerNum, err := read("Owner #:")
	if err != nil {
		return err
	}

	owner, err := strconv.Atoi(ownerNum)
	if err != nil {
		return err
	}

	ownerId := owners[owner-1]

	c := &Config{
		User:    user,
		Pass:    pass,
		OwnerId: ownerId,
	}

	json, err := json.Marshal(c)
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.Write(json)
	if err != nil {
		return err
	}

	fmt.Println("\nConfiguration saved successfully")

	return nil
}
