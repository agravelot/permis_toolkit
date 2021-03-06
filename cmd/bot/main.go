package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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
		return []ornikar.InstructorNextLessonsInterval{}, fmt.Errorf("failed to get current working directory: %w", err)
	}

	if _, err := os.Stat(cwd + "/" + databaseFileName); err == nil {
		jsonFile, err := os.Open(databaseFileName)
		if err != nil {
			return []ornikar.InstructorNextLessonsInterval{}, fmt.Errorf("unable read database file: %w", err)
		}
		defer jsonFile.Close()
		var localLessons []ornikar.InstructorNextLessonsInterval

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return []ornikar.InstructorNextLessonsInterval{}, fmt.Errorf("failed to read database file: %w", err)
		}

		json.Unmarshal(byteValue, &localLessons)

		return localLessons, nil
	}

	return []ornikar.InstructorNextLessonsInterval{}, nil
}

func writeDatabase(lessons []ornikar.InstructorNextLessonsInterval) error {
	file, err := json.MarshalIndent(lessons, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal lessons: %w", err)
	}

	err = ioutil.WriteFile(databaseFileName, file, 0644)
	if err != nil {
		return fmt.Errorf("failed to write database: %w", err)
	}

	return nil
}

type Config struct {
	OrnikarEmail    string
	OrnikarPassword string
	DiscordToken    string
	InstructorID    int
}

func getConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("unable loading .env file")
	}

	instructorID, err := strconv.Atoi(os.Getenv("INSTRUCTOR_ID"))
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse INSTRUCTOR_ID: %w", err)
	}

	return Config{
		OrnikarEmail:    os.Getenv("ORNIKAR_EMAIL"),
		OrnikarPassword: os.Getenv("ORNIKAR_PASSWORD"),
		DiscordToken:    os.Getenv("DISCORD_TOKEN"),
		InstructorID:    instructorID,
	}, err
}

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	dg, err := discord.Start(config.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	var cookie string

	err = ornikar.Login(&cookie, config.OrnikarEmail, config.OrnikarPassword)
	if err != nil {
		log.Fatal(err)
	}

	err = run(&config, dg, &cookie)
	if err != nil {
		log.Println(err)
	}

	for range time.Tick(1 * time.Minute) {
		err := run(&config, dg, &cookie)
		if err != nil {
			log.Println(err)
		}
	}
}

func run(config *Config, dg *discordgo.Session, cookie *string) error {
	log.Println("Requesting new lessons...")

	lessons, err := ornikar.GetRemoteLessons(cookie, config.InstructorID)
	if err != nil {
		return fmt.Errorf("failed to get remote lessons: %w", err)
	}

	localLessons, err := getLocalLessons()
	if err != nil {
		return fmt.Errorf("failed to get local lessons: %w", err)
	}

	diff := getNewAvailableLessons(localLessons, lessons)

	if len(diff) == 0 {
		return nil
	}

	// Discord limit to 2000 characters
	for _, chunk := range discord.Chunk(diff, 40) {
		m, err := formatMessage(chunk)
		if err != nil {
			return fmt.Errorf("failed to format message: %w", err)
		}
		log.Println(m)
		err = discord.Notify(dg, m)
		if err != nil {
			return fmt.Errorf("failed to notify: %w", err)
		}
	}

	writeDatabase(lessons)
	return nil
}

func formatMessage(lessons []ornikar.InstructorNextLessonsInterval) (string, error) {
	var datesString string
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone: %w", err)
	}

	for _, l := range lessons {
		datesString += fmt.Sprintf("- **%s** \n", l.StartsAt.In(loc).Format("Monday 02 January 2006 15:04:05"))
	}

	return fmt.Sprintf("%d nouvelle sessions disponnible : \n%s \n \nLien : https://app.ornikar.com/planning @everyone", len(lessons), datesString), nil
}
