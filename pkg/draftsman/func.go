package draftsman

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	path2 "path"
	"path/filepath"
	"sync"
)

func ProjectInclude(c IncludeElement, g *sync.WaitGroup) {
	defer g.Done()
	element := c
	projectId := getProjectId(element.Project, &ProjectsListUrl)
	pathToFile := downloadConfig(projectId, element.File, element.Ref, &ConfigUrl)

	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		log.Fatalln("Config reading error:", err)
		return
	}

	fmt.Printf("Start parse file: %v\n", pathToFile)
	includeList := parseConfig(string(data))

	for _, include := range includeList {
		if include.Project != "" {

			for _, file := range include.File {
				y := IncludeElement{
					Project: include.Project,
					Ref:     include.Ref,
					File:    file,
				}
				node := Node{
					Me:    path2.Join(element.Project, element.File, element.Ref),
					Child: path2.Join(include.Project, file, include.Ref),
				}
				List = append(List, node)
				g.Add(1)
				go ProjectInclude(y, g)
			}
		}
		if include.Local != "" {
			y := IncludeElement{
				Project: element.Project,
				Id:      projectId,
				Ref:     element.Ref,
				File:    include.Local,
			}

			node := Node{
				Me:    path2.Join(element.Project, element.File, element.Ref),
				Child: path2.Join(element.Project, include.Local, element.Ref),
			}
			List = append(List, node)

			g.Add(1)
			LocalInclude(y, g)
		}
	}
}

func LocalInclude(c IncludeElement, g *sync.WaitGroup) {
	defer g.Done()
	element := c
	projectId := c.Id
	pathToFile := downloadConfig(projectId, element.File, element.Ref, &ConfigUrl)

	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		log.Fatalln("Config reading error:", err)
		return
	}
	fmt.Printf("Start parse file: %v\n", pathToFile)
	includeList := parseConfig(string(data))

	for _, include := range includeList {
		if include.Project != "" {

			for _, file := range include.File {
				y := IncludeElement{
					Project: include.Project,
					Ref:     include.Ref,
					File:    file,
				}
				node := Node{
					Me:    path2.Join(element.Project, element.File, element.Ref),
					Child: path2.Join(include.Project, file, include.Ref),
				}
				List = append(List, node)
				g.Add(1)
				go ProjectInclude(y, g)
			}
		}
		if include.Local != "" {
			y := IncludeElement{
				Project: element.Project,
				Id:      projectId,
				Ref:     element.Ref,
				File:    include.Local,
			}

			node := Node{
				Me:    path2.Join(element.Project, element.File, element.Ref),
				Child: path2.Join(element.Project, include.Local, element.Ref),
			}
			List = append(List, node)

			g.Add(1)
			LocalInclude(y, g)
		}
	}
}

func getProjectId(pathWithNamespace string, c *url.URL) string {
	projectsListUrl := *c
	projectsListUrl.RawQuery = fmt.Sprintf(projectsListUrl.RawQuery, pathWithNamespace)
	client := http.Client{}
	req, err := http.NewRequest("GET", projectsListUrl.String(), nil)
	req.Header.Add("PRIVATE-TOKEN", AppConfig.Token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: refactor
	id := gjson.Parse(string(body)).Array()[0].Get("id")
	if id.String() == "" {
		log.Fatalf("There is error, cannot find project ID: %v", projectsListUrl.String())
	}
	return id.String()
}

func downloadConfig(projectId string, file string, ref string, c *url.URL) string {
	dir, filename := filepath.Split(file)
	path := path2.Join(AppConfig.TmpDir, projectId, dir)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			log.Fatalln("Dir creating error:", err)
		}
	}
	path = path2.Join(path, filename)
	fmt.Println("Downloading: ", path)

	configUrl := *c
	switch ref {
	case "":
		configUrl.RawQuery = ""
	default:
		configUrl.RawQuery = configUrl.RawQuery + ref
	}

	urlEncodedFile := url.PathEscape(file)
	configUrl.Path = fmt.Sprintf(configUrl.Path, projectId, file)
	configUrl.RawPath = fmt.Sprintf(configUrl.RawPath, projectId, urlEncodedFile)

	out, err := os.Create(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	client := http.Client{}
	req, err := http.NewRequest("GET", configUrl.String(), nil)
	req.Header.Add("PRIVATE-TOKEN", AppConfig.Token)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status: %s, url: %s", resp.Status, configUrl.String())
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return path
}

func parseConfig(file string) []IncludeLocation {
	config := GitlabCiConfig{}
	err := yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		log.Fatalf("parsing error: %v\n", err)
	}
	return config.Include
}

func (g Graph) Generate() {
	var b bytes.Buffer

	b.WriteString("```mermaid\ngraph TD;\n")
	for i := range List {
		b.WriteString(fmt.Sprintf("%s --> %s;\n", List[i].Me, List[i].Child))
	}
	b.WriteString("```")

	f, err := os.Create("graph.md")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = f.WriteString(b.String())
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (a *StringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}
