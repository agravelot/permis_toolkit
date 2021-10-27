package candilib

import (
	"fmt"
	"net/http"
	"strings"
)

func MagicLink(email string) error {
	url := "https://beta.interieur.gouv.fr/candilib/api/v2/auth/candidat/magic-link"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{"email":"%s"}`, email))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return err
	}

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("sec-ch-ua", "\";Not A Brand\";v=\"99\", \"Chromium\";v=\"94\"")
	req.Header.Add("DNT", "1")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.104 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("X-CLIENT-ID", "e42b21f7-c999-4814-bed0-69e046829e43.2.11.15-beta2.")
	// req.Header.Add("X-REQUEST-ID", "acbce3ba-aa9c-4b9f-bd77-eefffd0e747e")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Origin", "https://beta.interieur.gouv.fr")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "https://beta.interieur.gouv.fr/")
	req.Header.Add("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }

	return nil
}
