package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	infoUrl = "%s"
)

var trainsToCheck = map[string]string{
	"18:05": "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/S08409/9638",
	"18:10": "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/S01700/9742",
	"18:45": "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/S09818/9336",
	"19:00": "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/S09218/9544",
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

func getTrainInfo() (Response, error) {
	trainResponse := make(map[string]string)

	for trainDesc, trainUrl := range trainsToCheck {
		trainStatus, err := retrieveTrainStatus(trainUrl)
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

func retrieveTrainStatus(trainUrl string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(trainUrl))
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

func main() {
	// fmt.Println(getTrainInfo())
	lambda.Start(Handler)
}
