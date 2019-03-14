package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	trainCodeUrl   = "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/cercaNumeroTrenoTrenoAutocomplete/%s"
	trainStatusUrl = "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/%s/%s"
)

var trainsToCheck = map[string]string{
	"18:05": "9638",
	"18:10": "9742",
	"18:45": "9336",
	"19:00": "9544",
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type TrainInfo struct {
	CompRitardo []string
}

type TrainResponse struct {
	TrainStatus map[string]string
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	return getTrainInfo()
}

func getTrainCode(trainNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(trainCodeUrl, trainNumber))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	trainCodeResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`(?P<InfoToDrop>[0-9]{4} - [a-zA-Z ]+\|[0-9]{4}-)(?P<TrainCode>[A-Z0-9]{6})`)
	m := r.FindStringSubmatch(string(trainCodeResponse))
	n := r.SubexpNames()

	return mapSubexpNames(m, n)["TrainCode"], nil
}

func mapSubexpNames(m, n []string) map[string]string {
	matches, names := m[1:], n[1:]
	subexpNamesMap := make(map[string]string)

	for i, _ := range matches {
		subexpNamesMap[names[i]] = matches[i]
	}

	return subexpNamesMap
}

func retrieveTrainStatus(trainNumber string) (string, error) {
	trainCode, err := getTrainCode(trainNumber)

	resp, err := http.Get(fmt.Sprintf(trainStatusUrl, trainCode, trainNumber))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var trainInfo TrainInfo
	err = json.NewDecoder(resp.Body).Decode(&trainInfo)
	if err != nil {
		return "", err
	}

	return trainInfo.CompRitardo[0], nil
}

func getTrainInfo() (Response, error) {
	trainResponse := make(map[string]string)

	for trainDesc, trainNumber := range trainsToCheck {
		trainStatus, err := retrieveTrainStatus(trainNumber)
		if err != nil {
			fmt.Println(err)
			trainStatus = "Not available"
		}

		trainResponse[trainDesc] = trainStatus
	}

	out, err := json.MarshalIndent(trainResponse, "", " ")
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	return Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(out),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
