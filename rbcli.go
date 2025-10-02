package main

import (
   "bytes"
   "encoding/json"
   "fmt"
   "github.com/urfave/cli"
   "io"
   "io/ioutil"
   "mime/multipart"
   "net/http"
   "os"
   "os/user"
   "path"
   "strings"
   "strconv"
   "github.com/go-ldap/ldap/v3"
)

var baseRubberbandApiUri = os.Getenv("RUBBERBAND_URL") + "/api/upload"
var maxBodySyncSize = 400000000 // 400 MB

type response_struct struct {
   Status   string
   Url      string
   Errors   []string
   Basename string
}

var successCodes = map[int]bool {
   200: true,
   201: true,
   202: true,
}

func getUserEmail(username string) string {
   // translate system user into email address via LDAP
   // $ ldapsearch -H ldap://ldap.example.com -x -LLL '(uid=username)' mail
   // following https://cybernetist.com/2020/05/18/getting-started-with-go-ldap/
   ldapURL := "ldap://ldap.example.com"
   l, err := ldap.DialURL(ldapURL)
   if err != nil {
      panic(fmt.Sprintf("Failed to dial LDAP server ldap://ldap.example.com: %v", err))
   }
   defer l.Close()

   err = l.UnauthenticatedBind("")
   if err != nil {
      panic(fmt.Sprintf("Failed to bind LDAP: %v", err))
   }

   filter := fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(username))
   searchReq := ldap.NewSearchRequest("ou=People,ou=example,dc=example,dc=com", ldap.ScopeWholeSubtree, 0, 0, 0, false, filter, []string{"mail"}, []ldap.Control{})

   result, err := l.Search(searchReq)
   if err != nil {
      panic(fmt.Sprintf("Failed to query LDAP ou=People,ou=example,dc=example,dc=com: %w", err))
   }
   if len(result.Entries) != 1 {
      fmt.Println("LDAP query results:", result.Entries)
      panic(fmt.Sprintf("LDAP query returned %d instead of 1 result", len(result.Entries)))
   }
   email := result.Entries[0].GetAttributeValue("mail")
   if email == "" {
      fmt.Println("LDAP query result:", result.Entries[0])
      panic(fmt.Sprintf("LDAP query did not return email", len(result.Entries)))
   }

   return strings.ToLower(email)
}

func MakeRequest(Files []string, Tags string, Async bool, Expdate string) string {
   client := &http.Client{}
   useLdap, _ := strconv.ParseBool(os.Getenv("RBCLI_USE_LDAP"))
   var userEmail string
   if useLdap {
     // get system user
     usr, _ := user.Current()
     userEmail = getUserEmail(usr.Username)
   } else {
     userEmail = "ghost"
   }

   // add files
   bodyBuf := &bytes.Buffer{}
   bodyWriter := multipart.NewWriter(bodyBuf)

   for id, filename := range Files {
      name := fmt.Sprintf("file%d", id)
      fileWriter, err := bodyWriter.CreateFormFile(name, path.Base(filename))
      check(err)

      // open file handle
      fh, err := os.Open(filename)
      check(err)

      // iocopy
      _, err = io.Copy(fileWriter, fh)
      check(err)
   }

   bodyWriter.WriteField("tags", Tags)
   if Expdate != "" {
      bodyWriter.WriteField("expirationdate", Expdate)
   }

   contentType := bodyWriter.FormDataContentType()

   // turn on async for large uploads to prevent timeouts
   if bodyBuf.Len() >= maxBodySyncSize {
      Async = true
   }

   bodyWriter.Close()

   url := baseRubberbandApiUri
   if Async {
      url = url + "/async"
   }

   // make the actual request
   request, err := http.NewRequest("PUT", url, bodyBuf)
   check(err)

   request.Header.Add("Content-Type", contentType)
   request.Header.Add("X-Api-Token", os.Getenv("RUBBERBAND_API_KEY"))
   request.Header.Add("X-Forwarded-Email", userEmail)
   response, err := client.Do(request)
   check(err)

   defer response.Body.Close()

   // if non-successful return code, return the reason
   if _, ok := successCodes[response.StatusCode]; ok {
      resp := []response_struct{}

      json.NewDecoder(response.Body).Decode(&resp)
      response.Body.Close()

      statusresponse := ""
      for _, t := range resp {
         if t.Basename != "" {
            statusresponse += fmt.Sprintf("Bundle %s:\n", t.Basename)
         }
         status := strings.ToLower(t.Status)
         if status == "success" {
            statusresponse += fmt.Sprintf("Files successfully uploaded. Browse them here:\n%s", t.Url)
         } else if status == "found" {
            statusresponse += fmt.Sprintf("Files ignored because they were already archived. Browse them here:\n%s", t.Url)
         } else if status == "queued" {
            statusresponse += fmt.Sprintf("Files queued for uploading. Check your inbox for more information.")
         } else {
            statusresponse += fmt.Sprintf("Failed to upload.\nStatus: %s\nErrors:\n%s.", t.Status, strings.Join(t.Errors, "\n"))
         }
         statusresponse += fmt.Sprintf("\n")
      }
      return statusresponse
   } else {
      content, err := ioutil.ReadAll(response.Body)
      check(err)
      return fmt.Sprintf("Rubberband replied with %d:\n%s", response.StatusCode, content)
   }
}

func UploadFiles(FileArray []string, Tags string, Async bool, Expdate string) string {

   FileCount := len(FileArray)

   if FileCount == 0 {
      return fmt.Sprintf("Rubberband needs output files (.set, .out, .err, .meta, .solu). You provided %d. Aborting.", FileCount)
   } else {
      // TODO: add validation for ensuring each file is only listed once as an arguement
      // files listed more than once as args will cause rbcli to fail silently
      result := MakeRequest(FileArray, Tags, Async, Expdate)
      return result
   }
}

func check(err error) {
   if err != nil {
      panic(err)
   }
}

func main() {
   var async bool
   var tags string
   var expdate string

   app := cli.NewApp()
   app.Name = "rubberband-cli"
   app.Usage = "the rubberband command line client"
   app.Version = "1.0.0"

   app.Flags = []cli.Flag{
      cli.BoolFlag{
         Name:        "async",
         Usage:       "Run this command asynchronously. Receive the results in an email.",
         Destination: &async,
      },
      cli.StringFlag{
         Name:        "tags",
         Usage:       "A comma-separated list of tags to associate with the uploaded run.",
         Destination: &tags,
      },
      cli.StringFlag{
         Name:        "e",
         Usage:       "The expirationdate, should be provided in form \"2017-Aug-24\".",
         Destination: &expdate,
      },
   }

   app.Commands = []cli.Command{
      {
         Name:    "upload",
         Aliases: []string{"up"},
         Usage:   "Upload a list of related output files for parsing and processing.",
         Action: func(c *cli.Context) error {
            res := UploadFiles(c.Args(), tags, async, expdate)
            fmt.Println(res)
            return nil
         },
      },
   }
   app.Run(os.Args)
}
