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
	"os/exec"
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "saves certificates to parameter store",
	Long:  `saves certificates to parameter store`,
	Run:   save,
}

func mapViperToParameters() *types.SaveParameters {
	return &types.SaveParameters{
		DomainName:  viper.GetString("domain-name"),
		DomainEmail: viper.GetString("domain-email"),
		KeyId:       viper.GetString("key-id"),
		Path:        viper.GetString("pstore-path"),
		ValidDays:   viper.GetInt("valid-days"),
	}
}

func createFolderIfNotExist(folderPath string) error {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return os.MkdirAll(folderPath, os.ModePerm)
	}
	return nil
}

func getFiles(folderPath string) []os.FileInfo {

	f, err := os.Open(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	files, err := f.Readdir(-1)
	return files
}

func generateCert(domainName string, domainEmail string) {
	cmd := exec.Command("certbot", "certonly", "--standalone", "-d", domainName, "--email", domainEmail, "-n", "--agree-tos", "--expand")
	out, _ := cmd.CombinedOutput()
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to generate cert %s\n%s\n", err, string(out))
	}
	log.Printf("%s\n", string(out))
}

func save(_ *cobra.Command, _ []string) {
	param := mapViperToParameters()
	paramStorePath := param.Path
	validDays := float64(param.ValidDays)
	domainName := param.DomainName
	domainEmail := param.DomainEmail
	keyId := param.KeyId
	folderPath := fmt.Sprintf("/etc/letsencrypt/live/%s", domainName)
	session, err := aws.NewSession()
	if err != nil {
		log.Fatal(err)
		return
	}

	ssmSession := session.NewSsmSession()
	if exists, parameters := ssmSession.Exists(paramStorePath, validDays); exists {
		err := createFolderIfNotExist(folderPath)
		if err != nil {
			log.Fatalf("fail to create folder %q because %s", folderPath, err.Error())
		}
		ssmSession.Restore(parameters, folderPath)
		return
	}

	generateCert(domainName, domainEmail)

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	files := getFiles(folderPath)
	for _, file := range files {
		fileName := file.Name()
		filePath := fmt.Sprintf("%s/%s", folderPath, fileName)

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		ssmSession.Save(fileName, string(content), keyId, paramStorePath)
	}
}

func init() {
	rootCmd.AddCommand(saveCmd)
	saveCmd.PersistentFlags().String("domain-name", "", "--domain-name app.xyz.com")
	saveCmd.PersistentFlags().String("domain-email", "", "--domain-email ibu@xyz.com")
	saveCmd.PersistentFlags().String("key-id", "", "--key-id alias/aws/ssm")
	saveCmd.PersistentFlags().String("pstore-path", "", "--pstore-path /allEnvs/DevTest/ssl")
	saveCmd.PersistentFlags().Int("valid-days", 60, "--valid-days 30")
	err := viper.BindPFlags(saveCmd.PersistentFlags())

	if err != nil {
		log.Fatalf("fail to bind command arguments\n %s", err.Error())
	}
}
