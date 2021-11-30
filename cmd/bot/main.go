package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/agravelot/permis_toolkit/candilib"
	"github.com/agravelot/permis_toolkit/cron"
	"github.com/agravelot/permis_toolkit/discord"
	"github.com/agravelot/permis_toolkit/ornikar"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	databaseFileName = "database.json"
)

func getNewAvailableLessons(localLessons, remoteLessons []ornikar.InstructorNextLessonsInterval) []ornikar.InstructorNextLessonsInterval {
	// if reflect.DeepEqual(localLessons, remoteLessons) {
	// 	return []InstructorNextLessonsInterval{}
	// }

	newLessons := []ornikar.InstructorNextLessonsInterval{}

	for _, rl := range remoteLessons {
		isNew := true
		if rl.StartsAt.Before(time.Now()) {
			continue // Ignore old sessions
		}
		// Search if already fetched in previous run.
		for _, ll := range localLessons {
			if ll.ID == rl.ID {
				isNew = false
				// TODO Remove it from localLessons
			}
		}

		if isNew {
			newLessons = append(newLessons, rl)
		}
	}

	return newLessons
}

func getLocalLessons() ([]ornikar.InstructorNextLessonsInterval, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return []ornikar.InstructorNextLessonsInterval{}, err
	}

	if _, err := os.Stat(cwd + "/" + databaseFileName); err == nil {
		jsonFile, err := os.Open(databaseFileName)
		if err != nil {
			return []ornikar.InstructorNextLessonsInterval{}, err
		}
		defer jsonFile.Close()
		var localLessons []ornikar.InstructorNextLessonsInterval

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return []ornikar.InstructorNextLessonsInterval{}, err
		}

		json.Unmarshal(byteValue, &localLessons)

		return localLessons, nil
	}

	return []ornikar.InstructorNextLessonsInterval{}, nil
}

func writeDatabase(lessons []ornikar.InstructorNextLessonsInterval) error {
	file, err := json.MarshalIndent(lessons, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(databaseFileName, file, 0644)
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	OrnikarEmail    string
	CandilibEmail   string
	OrnikarPassword string
	DiscordToken    string
	InstructorID    int
}

func getConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	instructorID, err := strconv.Atoi(os.Getenv("INSTRUCTOR_ID"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		CandilibEmail:   os.Getenv("CANDILIB_EMAIL"),
		OrnikarEmail:    os.Getenv("ORNIKAR_EMAIL"),
		OrnikarPassword: os.Getenv("ORNIKAR_PASSWORD"),
		DiscordToken:    os.Getenv("DISCORD_TOKEN"),
		InstructorID:    instructorID,
	}, err
}

func main() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	dg, err := discord.Start(config.DiscordToken)
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	var cookie string

	err = ornikar.Login(&cookie, config.OrnikarEmail, config.OrnikarPassword)
	if err != nil {
		panic(err)
	}

	go cron.Run(func() error {
		return candilib.MagicLink(config.CandilibEmail)
	})

	run(&config, dg, &cookie)

	for range time.Tick(time.Second * 60) {
		run(&config, dg, &cookie)
	}
}

func run(config *Config, dg *discordgo.Session, cookie *string) {
	log.Println("Requestiong new lessons...")

	lessons, err := ornikar.GetRemoteLessons(cookie)
	if err != nil {
		panic(err)
	}

	localLessons, err := getLocalLessons()
	if err != nil {
		panic(err)
	}

	diff := getNewAvailableLessons(localLessons, lessons)

	if len(diff) == 0 {
		return
	}

	m, err := formatMessage(diff)
	if err != nil {
		panic(err)
	}
	log.Println(m)
	err = discord.Notify(dg, m)
	if err != nil {
		panic(err)
	}

	writeDatabase(lessons)
}

func formatMessage(lessons []ornikar.InstructorNextLessonsInterval) (string, error) {
	var datesString string
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return "", err
	}

	for _, l := range lessons {
		datesString += fmt.Sprintf("- **%s** \n", l.StartsAt.In(loc).Format("Monday 02 January 2006 15:04:05"))
	}

	return fmt.Sprintf("%d nouvelle sessions disponnible : \n%s \n \nLien : https://app.ornikar.com/planning @everyone", len(lessons), datesString), nil
}
