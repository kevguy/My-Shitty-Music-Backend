package fcm

import (
	"fmt"
	"log"
	"path/filepath"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"

	"google.golang.org/api/option"
)

type FcmClient struct {
	client *messaging.Client
	ctx    context.Context
}

func (fcmClient *FcmClient) SubscribeToBroadcastTopic(token string) {
	registrationTokens := []string{
		token,
		// "YOUR_REGISTRATION_TOKEN_1",
		// ...
		// "YOUR_REGISTRATION_TOKEN_n",
	}

	// Subscribe the devices corresponding to the registration tokens to the
	// topic.
	response, err := fcmClient.client.SubscribeToTopic(fcmClient.ctx, registrationTokens, "broadcast")
	if err != nil {
		log.Fatalln(err)
	}

	// See the TopicManagementResponse reference documentation
	// for the contents of response.
	fmt.Println(response.SuccessCount, "tokens were subscribed successfully")
}

func (fcmClient *FcmClient) UnsubscribeFromBroadcastTopic(token string) {
	registrationTokens := []string{
		token,
		// "YOUR_REGISTRATION_TOKEN_1",
		// ...
		// "YOUR_REGISTRATION_TOKEN_n",
	}

	// Unsubscribe the devices corresponding to the registration tokens to the
	// topic.
	response, err := fcmClient.client.UnsubscribeFromTopic(fcmClient.ctx, registrationTokens, "broadcast")
	if err != nil {
		log.Fatalln(err)
	}

	// See the TopicManagementResponse reference documentation
	// for the contents of response.
	fmt.Println(response.SuccessCount, "tokens were subscribed successfully")
}

func (fcmClient *FcmClient) BroadcastMessage(title string, body string) {
	fmt.Println("BroadcastMessage")
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Icon:  "https://raw.githubusercontent.com/kevguy/My-Shitty-Music-Frontend/master/public/img/icons/icon-72x72.png",
			},
		},
		Topic: "broadcast",
	}

	response, err := fcmClient.client.Send(fcmClient.ctx, message)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(response)
	}
}

// BroadcastHello says Hello to everyone
func (fcmClient *FcmClient) BroadcastHello() {
	// oneHour := time.Duration(1) * time.Hour
	// badge := 42
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "Hello World",
			Body:  "Just dropping by and say a fuckin hello",
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: "Hello World",
				Body:  "Just dropping by and say a fuckin hello",
				Icon:  "https://raw.githubusercontent.com/kevguy/My-Shitty-Music-Frontend/master/public/img/icons/icon-72x72.png",
			},
		},
		// Android: &messaging.AndroidConfig{
		// 	TTL: &oneHour,
		// 	Notification: &messaging.AndroidNotification{
		// 		Icon:  "stock_ticker_update",
		// 		Color: "#f45342",
		// 	},
		// },
		// APNS: &messaging.APNSConfig{
		// 	Payload: &messaging.APNSPayload{
		// 		Aps: &messaging.Aps{
		// 			Badge: &badge,
		// 		},
		// 	},
		// },
		Topic: "broadcast",
	}
	response, err := fcmClient.client.Send(fcmClient.ctx, message)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(response)
	}

	// message = &messaging.Message{
	// 	Notification: &messaging.Notification{
	// 		Title: "Hello World",
	// 		Body:  "Just dropping by and say a hello",
	// 	},
	// 	Android: &messaging.AndroidConfig{
	// 		// TTL: &oneHour,
	// 		Notification: &messaging.AndroidNotification{
	// 			Icon:  "stock_ticker_update",
	// 			Color: "#f45342",
	// 		},
	// 	},
	// 	// APNS: &messaging.APNSConfig{
	// 	// 	Payload: &messaging.APNSPayload{
	// 	// 		Aps: &messaging.Aps{
	// 	// 			Badge: &badge,
	// 	// 		},
	// 	// 	},
	// 	// },
	// 	Token: "djTi9Lj89o4:APA91bGbK5OO9AHsiY3_C0lWgMrOzklxr8ywcz4qCn4j9a2KSV9cASdnDlfe9rBHDW-Dm0ItvKEyn5Z6i9g4GZpdv0L3gcNq9I5RUaSdXVm3AqyrmuHqF1FOFF59vRETU67KKzkfplfu",
	// }
	//
	// response, err = fcmClient.client.Send(fcmClient.ctx, message)
	// if err != nil {
	// 	log.Fatalln(err)
	// } else {
	// 	fmt.Println(response)
	// }
}

func InitFcmClient() *FcmClient {
	configPath, err := filepath.Abs("./my-shitty-music-firebase-adminsdk-mxvod-71bff034b8.json")
	if err != nil {
		log.Fatalf("error getting config: %v\n", err)
	}
	opt := option.WithCredentialsFile(configPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	fcmClient := &FcmClient{
		ctx:    ctx,
		client: client,
	}

	return fcmClient
}
