package notifications

// import (
// 	"context"
// 	"log"

// 	firebase "firebase.google.com/go"
// 	"firebase.google.com/go/messaging"
// 	"github.com/puoklam/chat-app-backend/env"
// 	"google.golang.org/api/option"
// )

// var app *firebase.App

// var client *messaging.Client

// func init() {
// 	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(env.FIREBASE_CRED_PATH))
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	client, err = app.Messaging(context.Background())
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }

// func Send(ctx context.Context, data map[string]string, token string) (string, error) {
// 	return client.Send(ctx, &messaging.Message{
// 		Data:  data,
// 		Token: token,
// 	})
// }

// func SendMultiCast(ctx context.Context, data map[string]string, tokens []string) (*messaging.BatchResponse, error) {
// 	return client.SendMulticast(ctx, &messaging.MulticastMessage{
// 		Data:   data,
// 		Tokens: tokens,
// 	})
// }

// // // Import the functions you need from the SDKs you need
// // import { initializeApp } from "firebase/app";
// // import { getAnalytics } from "firebase/analytics";
// // // TODO: Add SDKs for Firebase products that you want to use
// // // https://firebase.google.com/docs/web/setup#available-libraries

// // // Your web app's Firebase configuration
// // // For Firebase JS SDK v7.20.0 and later, measurementId is optional
// // const firebaseConfig = {
// //   apiKey: "AIzaSyDh1f4Zyqdh76BijMRjcSpPH_JQmN7_fkg",
// //   authDomain: "im-push-notification.firebaseapp.com",
// //   projectId: "im-push-notification",
// //   storageBucket: "im-push-notification.appspot.com",
// //   messagingSenderId: "429729531661",
// //   appId: "1:429729531661:web:742e194242404f21d05244",
// //   measurementId: "G-YT18YSV23R"
// // };

// // // Initialize Firebase
// // const app = initializeApp(firebaseConfig);
// // const analytics = getAnalytics(app);
