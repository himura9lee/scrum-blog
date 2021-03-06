package md5view

import (
	"blog/internal/process"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	ShortSeqTime string = "20060102"
	LongSeqTime string = "20060102150405"
	LongSplitTime string = "2006-01-02 15:04:05"
)

func EditConfigJS(file string, key, value string) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	pat := fmt.Sprintf("%s: '[^']*'", key)
	re, _ := regexp.Compile(pat)
	cp := re.ReplaceAll(raw, []byte(fmt.Sprintf("%s: '%s'", key, value)))
	ioutil.WriteFile(file, cp, 0666)
	return nil
}

type FrontMatter struct {
	Title string `yaml:"title,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
	Categories []string `yaml:"categories,omitempty"`
	Publish bool `yaml:"publish,omitempty"`
	Date string `yaml:"date,omitempty"`
	Passwd []string `yaml:"keys,omitempty"`
}

type VuePressDoc struct {
	FrontMatter *FrontMatter
	Doc string
}

func LoadVuePressDoc(path string) (*VuePressDoc, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lo := strings.Index(string(raw), "---\n")
	ro := strings.LastIndex(string(raw), "---")
	if ro >= lo {
		return nil, errors.New("illegal format")
	}
	fm := new(FrontMatter)
	err = yaml.Unmarshal(raw[lo+4:ro], fm)
	if err != nil {
		return nil, err
	}
	return &VuePressDoc{
		FrontMatter: fm,
		Doc:         string(raw[ro+4:]),
	}, nil
}

func EditVuePressDoc(path, key string, value ...interface{}) error {
	vpd, err := LoadVuePressDoc(path)
	if err != nil {
		return err
	}
	switch key {
	case "title":
		vpd.FrontMatter.Title = value[0].(string)
	case "tags":
		for i := 0; i < len(vpd.FrontMatter.Tags); i ++ {
			if vpd.FrontMatter.Tags[i] == value[1].(string) {
				vpd.FrontMatter.Tags[i] = value[0].(string)
				if value[0].(string) == "" {
					for j := i; j < len(vpd.FrontMatter.Tags) - 1; j ++ {
						vpd.FrontMatter.Tags[j] = vpd.FrontMatter.Tags[j+1]
					}
					vpd.FrontMatter.Tags = vpd.FrontMatter.Tags[:len(vpd.FrontMatter.Tags)-1]
				}
				break
			}
		}
	case "categories":
		for i := 0; i < len(vpd.FrontMatter.Categories); i ++ {
			if vpd.FrontMatter.Categories[i] == value[1].(string) {
				vpd.FrontMatter.Categories[i] = value[0].(string)
				if value[0].(string) == "" {
					for j := i; j < len(vpd.FrontMatter.Categories) - 1; j ++ {
						vpd.FrontMatter.Categories[j] = vpd.FrontMatter.Categories[j+1]
					}
					vpd.FrontMatter.Categories = vpd.FrontMatter.Categories[:len(vpd.FrontMatter.Categories)-1]
				}
				break
			}
		}
	case "publish":
		vpd.FrontMatter.Publish = value[0].(bool)
	case "date":
		vpd.FrontMatter.Date = value[0].(time.Time).Format(LongSplitTime)
	}
	return nil
}

func (vpd *VuePressDoc) String() string {
	fm, _ := yaml.Marshal(vpd.FrontMatter)
	return fmt.Sprintf("---\n%s---\n%s", string(fm), vpd.Doc)
}

func YarnBuild(path string) {
	p := process.NewProcess(exec.Command("bash", path))
	err := p.Start()
	if err != nil {
		return
	}
	err = p.Wait()
	logrus.Info("rebuild vuePress result", "cmd", p.Cmd(), "out", p.Stdout(), "err", p.Stderr())
	if err != nil {
		return
	}
}