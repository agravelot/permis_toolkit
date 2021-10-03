package ornikar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const ornikarUrl = "https://www.ornikar.com/connexion"
const loginButtonSelector = "#app-boilerplate-root-element > div > div.styles_AppPageWithoutShell__22P5T > div.styles_AuthenticationLayout__381Bv > div.styles_Container__cW10h > div.kitt_Container_XI20R.kitt_DepthHigh_2cpa8.styles_Card__9TFUm > div > div.styles_Form__3rHXh > div > form > div.auth_SignInSubmit_3mDF7 > button"
const homeSelector = "#app-boilerplate-root-element > div > div:nth-child(2) > div > div.styles_Content__2ZZRg > div:nth-child(2) > div.styles_GridLayout__1Sg3D > div.styles_Row__3bxvG > div.styles_Column__YfouA.styles_Span2__-FcHP > div.styles_Container__1Xn_M.styles_ContainerWithActions__1mSJa.styles_ContainerWithActionsOnColumn__cGdPZ.styles_SectionHeader__3ZvjM > div.styles_Title__6PgIA > h4"

func Login(email, password string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true), // To display the browser
		chromedp.Flag("enable-automation", false),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	)
	initCtx, initCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer initCancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		initCtx, chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	var cookie string
	if err := chromedp.Run(ctx, loginTasks(&ctx, ornikarUrl, &cookie, email, password)); err != nil {
		return "", err
	}

	cookie = "lwaat=aze"

	return cookie, nil
}

// func runWithTimeOut(ctx *context.Context, timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
// 	return func(ctx context.Context) error {
// 		timeoutContext, cancel := context.WithTimeout(ctx, timeout*time.Second)
// 		defer cancel()
// 		return tasks.Do(timeoutContext)
// 	}
// }

func loginTasks(ctx *context.Context, urlstr string, cookie *string, email, password string) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate(urlstr),

		chromedp.WaitVisible("#email"),
		chromedp.SendKeys(`//input[@name="email"]`, email),
		chromedp.SendKeys(`//input[@name="password"]`, password),
		chromedp.WaitVisible(loginButtonSelector),
		chromedp.Click(loginButtonSelector, chromedp.NodeVisible),
		chromedp.WaitVisible(homeSelector),

		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}

			for _, c := range cookies {
				if c.Name == "lwaat" {
					*cookie = fmt.Sprintf("%s=%s", c.Name, c.Value)
					return nil
				}
			}

			return errors.New("unable to get lwaat cookie")
		}),
	}
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Typename  string  `json:"__typename"`
}

type MeetingPoint struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Timezone string   `json:"timezone"`
	Location Location `json:"location"`
	Typename string   `json:"__typename"`
}

type InstructorNextLessonsInterval struct {
	ID           string       `json:"id"`
	StartsAt     time.Time    `json:"startsAt"`
	EndsAt       time.Time    `json:"endsAt"`
	MeetingPoint MeetingPoint `json:"meetingPoint"`
	Typename     string       `json:"__typename"`
}

type LessonsResponse struct {
	Data struct {
		InstructorNextLessonsInterval []InstructorNextLessonsInterval `json:"instructorNextLessonsInterval"`
	} `json:"data"`
}

type OrnikarError struct {
	StatusCode int
	Body       string
}

func (e *OrnikarError) Error() string {
	return fmt.Sprintf("ornikar api call failed, status code : %d, body : %s", e.StatusCode, string(e.Body))
}

type UnauthenticatedOrnikarError struct {
	OrnikarError
}

func (e *UnauthenticatedOrnikarError) Error() string {
	return "unauthenticated"
}

func GetRemoteLessons(cookie string) ([]InstructorNextLessonsInterval, error) {
	url := "https://app-gateway.ornikar.com/graphql"
	method := "POST"
	instructorID := 1679645

	now, err := json.Marshal(time.Now())
	if err != nil {
		return []InstructorNextLessonsInterval{}, err
	}

	// TODO Add more config for params
	payload := strings.NewReader(fmt.Sprintf(`{
		"operationName": "InstructorNextLessonsIntervalQuery",
		"variables": {
			"input": {
				"instructorId": "%d",
				"from": %s,
				"interval": 300,
				"gearbox": "MANUAL"
			}
		},
		"query": "query InstructorNextLessonsIntervalQuery($input: InstructorNextLessonsIntervalInput!) {instructorNextLessonsInterval(input: $input) {    id    startsAt    endsAt    meetingPoint {      id      name      timezone      location {        latitude        longitude        __typename      }      __typename    }    __typename  }}"
	}`, instructorID, now))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return []InstructorNextLessonsInterval{}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "fr,en-US;q=0.7,en;q=0.3")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("apollographql-client-name", "web")
	req.Header.Add("apollographql-client-version", "2.96.0")
	req.Header.Add("Origin", "https://app.ornikar.com")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Referer", "https://app.ornikar.com/")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", cookie)
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("TE", "trailers")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return []InstructorNextLessonsInterval{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return []InstructorNextLessonsInterval{}, err
	}

	// TODO Cookie as pinter and renew here
	if res.StatusCode == http.StatusBadRequest && strings.Contains(string(body), "UNAUTHENTICATED") {
		return []InstructorNextLessonsInterval{}, &UnauthenticatedOrnikarError{
			OrnikarError: OrnikarError{
				StatusCode: res.StatusCode,
				Body:       string(body),
			},
		}
	}

	if res.StatusCode != http.StatusOK {
		return []InstructorNextLessonsInterval{}, &OrnikarError{
			StatusCode: res.StatusCode,
			Body:       string(body),
		}
	}

	var lessons LessonsResponse
	if err := json.Unmarshal(body, &lessons); err != nil {
		return []InstructorNextLessonsInterval{}, err
	}

	return lessons.Data.InstructorNextLessonsInterval, nil
}
