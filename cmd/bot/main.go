package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

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
	OrnikarPassword string
	DiscordToken    string
}

func getConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		OrnikarEmail:    os.Getenv("ORNIKAR_EMAIL"),
		OrnikarPassword: os.Getenv("ORNIKAR_PASSWORD"),
		DiscordToken:    os.Getenv("DISCORD_TOKEN"),
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

	for _, l := range diff {
		m := formatMessage(l.StartsAt)
		err := discord.Notify(dg, m)
		println(m)
		if err != nil {
			panic(err)
		}
	}

	writeDatabase(lessons)
}

func formatMessage(date time.Time) string {
	return "Nouvelle sessions disponnible : " + "**" + date.Format("02 January 2006 15:04:05") + "**" + "\nLien : https://app.ornikar.com/planning"
}
