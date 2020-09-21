/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/kahgeh/pscert/aws"
	"github.com/kahgeh/pscert/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "saves certificates to parameter store",
	Long: `saves certificates to parameter store`,
	Run: save,
}

func mapToConfig() (string, string, string, string){
	return viper.GetString("domain-name"),viper.GetString("folder-path"), viper.GetString("key-id"), viper.GetString("pstore-path")
}

func getFiles(folderPath string) []os.FileInfo{

	f, err := os.Open(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	files, err := f.Readdir(-1)
	return files
}

func save(cmd *cobra.Command, args []string) {
	_, folderPath, keyId, pstorePath := mapToConfig()
	session, err := aws.NewSession()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("folder path = %q", folderPath)

	if _, err := os.Stat(folderPath); os.IsNotExist(err){
		log.Fatal(err)
	}
	files := getFiles(folderPath)
	for _,file := range files {
		fileName := file.Name()
		filePath :=fmt.Sprintf("%s/%s", folderPath, fileName)

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		session.Save(fileName, string(content), keyId, pstorePath)
	}
}

func init() {
	rootCmd.AddCommand(saveCmd)
	saveCmd.PersistentFlags().String("domain-name", "", "--domain-name app.xyz.com")
	saveCmd.PersistentFlags().String("folder-path", "", "--folder-path /etc/letsencrypt/live/app.xyz.com")
	saveCmd.PersistentFlags().String("key-id", "", "--key-id alias/aws/ssm")
	saveCmd.PersistentFlags().String("pstore-path", "", "--pstore-path /allEnvs/DevTest/ssl")
	err := viper.BindPFlags(saveCmd.PersistentFlags())

	if err != nil {
		log.Printf("fail to bind command arguments\n %s", err.Error())
		os.Exit(int(types.ExitFail))
		return
	}
}
