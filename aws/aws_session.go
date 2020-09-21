package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/kahgeh/pscert/ctx"
	"log"
	"os"
	"strings"
	"time"
)

type Session struct {
	config aws.Config
}

type SsmSession struct {
	api *ssm.Client
}

func NewSession() (*Session, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	session := &Session{config: cfg}
	return session, nil
}

func (session *Session) NewSsmSession() *SsmSession {
	api := ssm.New(session.config)
	return &SsmSession{
		api: api,
	}
}
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func writeFile(content string, filePath string) {
	f, err := os.Create(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(content)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func (session *SsmSession) Exists(path string, validDays float64) (bool, []ssm.Parameter) {
	api := session.api
	log.Println("checking pstore parameter exist...")
	request := api.GetParametersByPathRequest(&ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	response, err := request.Send(ctx.GetContext())
	if err != nil {
		log.Printf("err %s\n", err.Error())
		return false, nil
	}
	log.Printf("pstore parameter count = %v \n", len(response.Parameters))
	if len(response.Parameters) < 1 {
		return false, nil
	}
	log.Printf(response.String())
	now := time.Now()
	today := date(now.Year(), now.Month(), now.Day())
	year, month, day := response.Parameters[0].LastModifiedDate.Date()
	modifiedDate := date(year, month, day)
	log.Printf("pstore parameter days %v \n", (today.Sub(modifiedDate).Hours() / 24))
	return (today.Sub(modifiedDate).Hours() / 24) < validDays, response.Parameters
}

func (session *SsmSession) Restore(parameters []ssm.Parameter, folderPath string) {
	for _, parameter := range parameters {
		content := *parameter.Value
		fullName := *parameter.Name
		parts := strings.Split(fullName, "/")
		name := parts[len(parts)-1]
		filePath := fmt.Sprintf("%s/%s", folderPath, name)
		writeFile(content, filePath)
		log.Printf("saved content with len=%v to %q", len(content), filePath)
	}
}

func (session *SsmSession) Save(name string, content string, keyId string, path string) {
	if strings.HasPrefix(name, ".") {
		log.Printf("skipping %q ", name)
		return
	}
	api := session.api
	tier := ssm.ParameterTierStandard
	if len(content) > 4000 {
		tier = ssm.ParameterTierAdvanced
	}
	fullName := fmt.Sprintf("%s/%s", path, name)
	request := api.PutParameterRequest(&ssm.PutParameterInput{
		Type:      ssm.ParameterTypeSecureString,
		KeyId:     aws.String(keyId),
		Name:      aws.String(fullName),
		Value:     aws.String(content),
		Tier:      tier,
		Overwrite: aws.Bool(true),
	})
	response, err := request.Send(ctx.GetContext())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("saved %q to parameter store with len=%v, version=%v\n", name, len(content), response.Version)
}
