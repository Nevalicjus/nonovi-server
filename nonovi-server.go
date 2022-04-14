package main

import "fmt"
import "os"
import "github.com/gin-gonic/gin"
import "log"
import "strings"
import "net/http"
import "time"
import "crypto/md5"
import "encoding/base64"
import "gopkg.in/yaml.v2"

var nnvs = nnvsCollection{}
var userhome string = makeHome()
var config = loadConfig(userhome + "/.config/nonovi-server/conf.yaml")

type nnvsCollection struct {
	Nnvs map[string]Nnv
}
type Nnv struct {
	Name  string
	Md5   string
	Board string
}

type Config struct {
	NnvsDirectory string `yaml:"nnvsdir"`
	Port string `yaml:"port"`
}

func loadConfig(fp string) (*Config) {
	config := &Config{}
	_, err := os.Stat(fp)
	if os.IsNotExist(err) == true {
		config.NnvsDirectory = "/.config/nonovi-server/nnvs/"
		config.Port = ":9020"
		return config
	}
    f, err := os.ReadFile(fp)
    if err != nil { log.Fatal(err) }
    err = yaml.Unmarshal([]byte(f), &config)
    if err != nil { log.Fatal(err) }
    return config
}

func ifExists(m nnvsCollection, i string) bool {
	for _, value := range m.Nnvs {
		if value.Name == i {
			return true
		}
	}
	return false
}

func readDirectory(root string) ([]string, error) {
    var files []string
    f, err := os.Open(root)
    if err != nil { return files, err }
    fileInfo, err := f.Readdir(-1)
    f.Close()
    if err != nil { return files, err }

    for _, file := range fileInfo {
        files = append(files, file.Name())
    }
    return files, nil
}

func makeHome() string {
	userhome, err := os.UserHomeDir()
	if err != nil { log.Fatal(err) }
	err = os.MkdirAll(userhome + "/.config/nonovi-server/nnvs/", os.ModePerm)
	if err != nil { log.Fatal(err) }
	return userhome
}

func get_nnv (c *gin.Context) {
	id := c.Param("id")
	fmt.Println("[NNV]", id, "was requested" )

	b := ifExists(nnvs, id)
	if b != true { fmt.Println("Index not found") }

	c.JSON(200, gin.H { "nnv": []string{nnvs.Nnvs[id].Md5, nnvs.Nnvs[id].Board} })
}

func get_nnvs (c *gin.Context) {
	fmt.Println("[NNV] nnvs were requested")

	l := make([][]string, 0)
	for _, value := range nnvs.Nnvs {
		nnv := []string{value.Name, string(value.Md5)}
		l = append(l, nnv)
	}
	c.JSON(200, gin.H { "nnvs": l })
}

func loadnnvs() {
	files, err := readDirectory(userhome + config.NnvsDirectory)
	if err != nil { log.Fatal(err) }
	for _, file := range files {
		name := strings.Split(file, ".")[0]
		content, err := os.ReadFile(userhome + config.NnvsDirectory + file)
		if err != nil { log.Fatal(err) }

		var hash = md5.Sum(content)
		hash_str := base64.StdEncoding.EncodeToString(hash[:])

		nnv := Nnv{name, hash_str, string(content)}

		nnvs.Nnvs[string(name)] = nnv 
		fmt.Println("[NNV]", name, "was loaded")
	} 
}


func main() {
	nnvs.Nnvs = make(map[string]Nnv)
	loadnnvs()
	
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/get_nnv/:id", get_nnv)
	router.GET("/get_nnvs", get_nnvs)

	server := &http.Server{
		Addr:           config.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	
	server.ListenAndServe()
}
