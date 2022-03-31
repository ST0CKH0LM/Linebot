package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mehanizm/airtable"
)

var spoint string

func checkerr(err error) {
	if err != nil {
		log.Println(err)
	}
}

type Webhook struct {
	Destination string           `json:"destination"`
	Events      []*linebot.Event `json:"events"`
}

type Item struct {
	Userid    string
	CName     string
	Phone     string
	Peoplenum string
	CLocation string
	Point     string
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	client := airtable.NewClient("keyeKStkouk4crjCU")
	table := client.GetTable("appLALGhpqftoORtP", "tblALpCXZI9OA9b0s")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewStaticCredentials("AKIAQXDY6EKNDCUYK3NY", "YpYbe0pXqfxpZqRCuGUhNgcatCPPQpIwdJZAZVQe", ""),
	})
	checkerr(err)
	svc := dynamodb.New(sess)
	bot, err := linebot.New(
		"c980189c232e59b6ddc28a1c8dee3765",
		"+Esrwm4l5O7Vg/4KY44RuYruWj5AfTmN18dlCQv/IUjcbS781HGs1w0vgC1tymZwHR2MtwyeYSTA7wJldMstGB37Cz9QWlESf+1YfXnuPAMOEXhvo4NcX1vn98LbEZDFSFCRMHSjBO/xOoYB1JWDegdB04t89/1O/w1cDnyilFU=",
	)
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"message:":"%s"}`+"\n", http.StatusText(http.StatusInternalServerError)),
		}, nil
	}
	log.Print(request.Headers)
	log.Print(request.Body)
	if !validateSignature("c980189c232e59b6ddc28a1c8dee3765", request.Headers["x-line-signature"], []byte(request.Body)) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", linebot.ErrInvalidSignature.Error()),
		}, nil
	}

	var webhook Webhook

	if err := json.Unmarshal([]byte(request.Body), &webhook); err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", http.StatusText(http.StatusBadRequest)),
		}, nil
	}
	for _, event := range webhook.Events {
		log.Println(spoint)
		tableName := "Line"
		result, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"Userid": {S: aws.String(event.Source.UserID)},
			},
		})
		log.Println("error")
		checkerr(err)
		getitem := Item{}
		err = dynamodbattribute.UnmarshalMap(result.Item, &getitem)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		log.Println("Get Item")
		log.Println(getitem.Userid)
		log.Println(getitem.Point)
		if event.Type == linebot.EventTypeMessage {
			switch m := event.Message.(type) {
			case *linebot.TextMessage:
				if m.Text == "ติดต่อเจ้าหน้าที่" {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("กรุณาระบุชื่อของท่าน")).Do()
					checkerr(err)
					spoint = "1"
					item := Item{
						Userid:    event.Source.UserID,
						CName:     "",
						Phone:     "",
						Peoplenum: "",
						CLocation: "",
						Point:     spoint,
					}
					Additem(item, svc)
				} else if event.Source.UserID == getitem.Userid && getitem.Point == "1" {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("กรุณาระบุเบอร์โทรของท่าน")).Do()
					checkerr(err)
					log.Println(m.Text)
					spoint = "2"
					input_1 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(spoint),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Point = :s"),
					}
					test_1, err := svc.UpdateItem(input_1)
					checkerr(err)
					log.Println(test_1)
					input_2 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(m.Text),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set CName = :s"),
					}
					test_2, err := svc.UpdateItem(input_2)
					checkerr(err)
					log.Println(test_2)
				} else if event.Source.UserID == getitem.Userid && getitem.Point == "2" {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("กรุณาระบุจำนวนผู้ที่เข้าร่วมงาน")).Do()
					checkerr(err)
					spoint = "3"
					input_1 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(spoint),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Point = :s"),
					}
					test_1, err := svc.UpdateItem(input_1)
					checkerr(err)
					log.Println(test_1)
					input_2 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(m.Text),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Phone = :s"),
					}
					test_2, err := svc.UpdateItem(input_2)
					checkerr(err)
					log.Println(test_2)
				} else if event.Source.UserID == getitem.Userid && getitem.Point == "3" {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("กรุณาระบุสถานที่ที่ต้องการจัดงาน")).Do()
					checkerr(err)
					spoint = "4"
					log.Println(m.Text)
					input_1 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(spoint),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Point = :s"),
					}
					test_1, err := svc.UpdateItem(input_1)
					checkerr(err)
					log.Println(test_1)
					input_2 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(m.Text),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Peoplenum = :s"),
					}
					test_2, err := svc.UpdateItem(input_2)
					checkerr(err)
					log.Println(test_2)
				} else if event.Source.UserID == getitem.Userid && getitem.Point == "4" {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("ขอบคุณสำหรับข้อมูลครับ")).Do()
					checkerr(err)
					spoint = "5"
					log.Println(m.Text)
					input_1 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(spoint),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set Point = :s"),
					}
					test_1, err := svc.UpdateItem(input_1)
					checkerr(err)
					log.Println(test_1)
					input_2 := &dynamodb.UpdateItemInput{
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":s": {
								S: aws.String(m.Text),
							},
						},
						TableName: aws.String(tableName),
						Key: map[string]*dynamodb.AttributeValue{
							"Userid": {
								S: aws.String(event.Source.UserID),
							},
						},
						ReturnValues:     aws.String("UPDATED_NEW"),
						UpdateExpression: aws.String("set CLocation = :s"),
					}
					test_2, err := svc.UpdateItem(input_2)
					checkerr(err)
					log.Println(test_2)
					recordsToSend := &airtable.Records{
						Records: []*airtable.Record{
							{
								Fields: map[string]interface{}{
									"Userid":    getitem.Userid,
									"Name":      getitem.CName,
									"Phone":     getitem.Phone,
									"peopleNum": getitem.Peoplenum,
									"location":  m.Text,
								},
							},
						},
					}
					receivedRecords, err := table.AddRecords(recordsToSend)
					checkerr(err)
					log.Println(receivedRecords)
				} else {
					_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("ข้อมูลผิดพลาด กรุณาติดต่อเจ้าหน้าที่ใหม่อีกครั้ง")).Do()
					checkerr(err)
				}
			}
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func validateSignature(channelSecret string, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	_, err = hash.Write(body)
	if err != nil {
		return false
	}
	return hmac.Equal(decoded, hash.Sum(nil))
}

func Additem(item Item, svc *dynamodb.DynamoDB) {
	av, err := dynamodbattribute.MarshalMap(item)
	checkerr(err)
	log.Println(av)
	tableName := "Line"
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	save, err := svc.PutItem(input)
	checkerr(err)
	log.Println(save)
}

func main() {
	lambda.Start(handler)
}
