package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/kahgeh/pscert/ctx"
	"github.com/kahgeh/pscert/types"
	"log"
	"os"
)


type Session struct {
	config aws.Config
}

func NewSession() (*Session, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	session := &Session{config: cfg}
	return session, nil
}

func (session *Session) Save(name string, content string, keyId string, path string){
	api := ssm.New(session.config)
	tier:= ssm.ParameterTierStandard
	if len(content) > 4000 {
		tier = ssm.ParameterTierAdvanced
	}
	fullName := fmt.Sprintf("%s/%s",path, name)
	request := api.PutParameterRequest(&ssm.PutParameterInput{
		Type: ssm.ParameterTypeSecureString,
		KeyId: aws.String(keyId),
		Name: aws.String(fullName),
		Value: aws.String(content),
		Tier: tier,
	})
	response,err:=request.Send(ctx.GetContext())
	if err != nil {
		log.Println(err.Error())
		log.Println()
		os.Exit(int(types.ExitFail))
	}
	log.Println(response.String())
}

