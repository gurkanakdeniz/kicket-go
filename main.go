package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	// "io/ioutil"
	// "encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/matryer/goscript"
	"github.com/satori/go.uuid"
)

type CreateRequest struct {
	Code string `json: code`
	Uuid string `json: uuid`
}

type RunRequest struct {
	Args string `json: args`
}

var router *gin.Engine

func main() {
	router = gin.Default()
	initRoutes()
	router.Run()
}

func initRoutes() {
	router.GET("/", jedi)
	router.GET("/example", example)
	router.POST("/create", create)
	router.GET("/run/:uuid", run)
	router.POST("/run/:uuid", run)

	router.OPTIONS("/options", verification)
}

func jedi(c *gin.Context) {
	// message, _ := c.GetQuery("m")
	c.String(http.StatusOK, "return of the jedi")
}

func example(c *gin.Context) {
	// 	exampleCode := `
	//   import (
	// 		"strings"
	// 	)
	//
	// 	func goscript(name string) (string, error) {
	// 		return "Hello " + strings.ToUpper(name), nil
	// 	}
	// `
	exampleCode := `// you shouldn't change this -> <func ex(args ...string) string> , < package main >
	// you can add import but you shouldn't delete current import
package main

import (
	"fmt"
	"os"
)

func ex(args ...string) string {
	response := args[1]
	return response
}`

	exampleRequest := `{
    "args" : "jedi"
  }`

	c.JSON(http.StatusOK, gin.H{
		"exampleCode":    exampleCode,
		"exampleRequest": exampleRequest,
	})
}

func create(c *gin.Context) {
	var createRequest CreateRequest
	c.BindJSON(&createRequest)
	fmt.Printf("body -> %v\n", createRequest)

	uuidStr := getUUID()
	createFolder(uuidStr)
	createFile(uuidStr)
	createCode(uuidStr, createRequest.Code)

	c.JSON(http.StatusOK, gin.H{
		"endpoint": uuidStr,
		"platform": "go",
	})
}

func createCode(uuidStr string, code string) {
	// fmt.Printf("uuidStr - code -> %s %s\n", uuidStr, code)
	f, err := os.OpenFile(getFilePath(uuidStr), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
		return
	}

	code = code + "\n" + `func main() {
		fmt.Print(ex(os.Args[:]...))
	}`

	_, err = fmt.Fprintln(f, code)
	if err != nil {
		panic(err)
	}
	f.Close()
}

func run(c *gin.Context) {
	var runRequest RunRequest
	c.BindJSON(&runRequest)

	uuidStr := c.Param("uuid")
	// code := getCode(uuidStr)

	log.Println("------jedi------")
	
	//for _, pair := range os.Environ() {
	//    fmt.Println(pair)
	//}

	dir := os.Getenv("GOROOT")
	// fmt.Println(dir)
	dir = dir + "/bin/go"

	cmd := exec.Command("go", "run", getFilePath(uuidStr), runRequest.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	log.Println("-----jedi------")

	// script := goscript.New(code)
	// defer script.Close()

	// response, err := script.Execute(runRequest.Args)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	response := string(out)
	log.Println(response)

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func getCode(uuidStr string) string {
	f, err := os.OpenFile(getFilePath(uuidStr), os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
		return ""
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	contents := buf.String()

	f.Close()

	return contents
}

func getFolderPath(uuidStr string) string {
	folderPath := filepath.Join(".", "codes/"+uuidStr)
	return folderPath
}

func getFilePath(uuidStr string) string {
	folderPath := getFolderPath(uuidStr)
	filePath := filepath.Join(folderPath, uuidStr+".go")
	return filePath
}

func createFolder(uuidStr string) {
	folderPath := getFolderPath(uuidStr)
	os.MkdirAll(folderPath, os.ModePerm)
}

func createFile(uuidStr string) {
	filePath := getFilePath(uuidStr)
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
		return
	}
	f.Close()
}

func getUUID() string {
	newUuid := uuid.NewV4()
	// newUuid := uuid.Must(uuid.NewV4())
	uuidStr := strings.Replace(newUuid.String(), "-", "", -1)
	return uuidStr
}

func verification(c *gin.Context) {

	if c.Request.Method == "OPTIONS" {
		// setup headers
		c.Header("Allow", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "origin, content-type, accept")
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
	}
}
