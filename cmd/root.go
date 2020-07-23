/*
Copyright © 2020 Jiang Le <smartynaoki@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type frontMatter struct {
	Date       time.Time
	Title      string
	Categories []string
	Tags       []string
}

type hugoContent struct {
	Meta frontMatter
	Body string
}

var cfgFile string
var farboxFilesPath string
var outputPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "farbox2hugo",
	Short: "转换Farbox内容到hugo",
	Long:  `这是一个用来转换Farbox内容到hugo的小工具`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := ioutil.ReadDir(farboxFilesPath)
		if err != nil {
			return err
		}
		for _, fn := range files {
			fm := frontMatter{}
			hc := hugoContent{}
			f, err := os.Open(path.Join(farboxFilesPath, fn.Name()))
			if err != nil {
				return err
			}
			defer f.Close()
			ln := 0
			scanner := bufio.NewScanner(f)
			bodyBuffer := bytes.Buffer{}
			for scanner.Scan() {
				l := scanner.Text()
				titleRe := regexp.MustCompile(`(?i)Title:\s*(.+)`)
				dateRe := regexp.MustCompile(`(?i)Date:\s*(.+)`)
				tagsRe := regexp.MustCompile(`(?i)Tags:\s*(.+)`)
				categoryRe := regexp.MustCompile(`(?i)Category:\s*(.+)`)
				notSupportedRe := regexp.MustCompile(`(?i)\w+:\s*.+`)
				if ln < 10 {
					if l != "---" {
						if m := titleRe.FindStringSubmatch(l); m != nil {
							fm.Title = m[1]
						}
						if m := dateRe.FindStringSubmatch(l); m != nil {
							var err error
							fm.Date, err = dateparse.ParseLocal(m[1])
							if err != nil {
								fmt.Println(err)
							}
						}
						if m := tagsRe.FindStringSubmatch(l); m != nil {
							fm.Tags = strings.Split(m[1], " ")
						}
						if m := categoryRe.FindStringSubmatch(l); m != nil {
							fm.Categories = strings.Split(m[1], " ")
						}
						if len(notSupportedRe.FindStringIndex(l)) == 0 {
							bodyBuffer.WriteString(l)
							bodyBuffer.WriteString("\n")
						}
					}
				} else {
					bodyBuffer.WriteString(l)
					bodyBuffer.WriteString("\n")
				}
				ln++
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			fnDateRe := regexp.MustCompile(`(?i)(\d{4}-\d{2}-\d{2}).?`)
			if fm.Date.IsZero() {
				if m := fnDateRe.FindStringSubmatch(fn.Name()); m != nil {
					fm.Date, err = dateparse.ParseLocal(m[1])
				}
			}
			if fm.Title == `` {
				fm.Title = strings.Split(fn.Name(), ".")[0]
			}
			hc.Meta = fm
			hc.Body = bodyBuffer.String()
			ofn := path.Join(outputPath, strings.Split(fn.Name(), ".")[0]+".md")
			of, err := os.OpenFile(ofn, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			defer of.Close()
			if err != nil {
				return err
			}
			w := bufio.NewWriter(of)
			w.WriteString("---\n")
			metaContent, _ := yaml.Marshal(&hc.Meta)
			w.Write(metaContent)
			w.WriteString("---\n")
			w.WriteString(hc.Body)
			w.Flush()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.farbox2hugo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&farboxFilesPath, "farbox", "f", "./posts", "Farbox内容文件路径")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "./output", "输出路径")
}
