package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/shuheiktgw/go-lambda-linebot/parser"
)

const Prefix = "tmb"

const (
	PingCommand = "ping"
)

var (
	PingTopicArn    = os.Getenv("PING_TOPIC_ARN")
	UnknownTopicArn = os.Getenv("UNKNOWN_TOPIC_ARN")
)

func proxyResponse(statusCode int, body string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
}

func dispatch(events []*linebot.Event) error {
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if strings.HasPrefix(message.Text, Prefix) {
					// Remove `tmb` prefix and leading and trailing spaces
					command := strings.TrimSpace(strings.TrimPrefix(message.Text, Prefix))
					switch command {
					case PingCommand:
						return publish(PingTopicArn, event)
					default:
						return publish(UnknownTopicArn, event)
					}
				}
			}
		}
	}

	return nil
}

func publish(arn string, event *linebot.Event) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := sns.New(sess)

	message, err := json.Marshal(event)
	if err != nil {
		return err
	}

	input := sns.PublishInput{Message: aws.String(string(message)), TopicArn: aws.String(arn)}
	_, err = client.Publish(&input)
	if err != nil {
		return err
	}

	return nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("tombot callback started, header: %s, body: %s%", request.Headers, request.Body)

	lineEvents, err := parser.ParseRequest(os.Getenv("CHANNEL_SECRET"), &request)
	if err != nil {
		fmt.Printf("error occurred while parsing request: %s", err)
		return *proxyResponse(http.StatusInternalServerError, fmt.Sprintf(`{"message": %s}`, err)), nil
	}

	err = dispatch(lineEvents)
	if err != nil {
		fmt.Printf("error occurred while dispatching request: %s", err)
		return *proxyResponse(http.StatusInternalServerError, fmt.Sprintf(`{"message": %s}`, err)), nil
	}

	return *proxyResponse(http.StatusOK, ""), nil
}

func main() {
	lambda.Start(handler)
}
