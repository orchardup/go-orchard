package main

import "fmt"
import "errors"
import "io/ioutil"
import "net/url"
import "net/http"
import "encoding/json"
import "code.google.com/p/gopass"

func main() {
  username, password := Prompt()
  token, err := GetToken(username, password)
  if err != nil {
    fmt.Printf("Error getting token:\n%s\n", err)
    return
  }
  fmt.Printf("token: %s\n", token)
}

func Prompt() (string, string) {
  var (
    username string
    password string
  )
  fmt.Print("Username you signed up for Orchard with: ")
  fmt.Scanln(&username)
  password, _ = gopass.GetPass("Password: ")
  return username, password
}

type AuthResponse struct {
  Token string
}

func GetToken(username string, password string) (string, error) {
  resp, err := http.PostForm("http://localdocker:8000/api/v1/signin",
    url.Values{"username": {username}, "password": {password}})

  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }

  if resp.StatusCode != 200 {
    return "", errors.New(fmt.Sprintf("erroneous API response: %s", body))
  }

  var authResponse AuthResponse
  if err := json.Unmarshal(body, &authResponse); err != nil {
    return "", err
  }

  return authResponse.Token, nil
}
