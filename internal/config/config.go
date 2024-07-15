package config

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type FirebaseApp struct {
	*firebase.App
}

func NewFirebaseApp(credentialFile string) (*FirebaseApp, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return &FirebaseApp{App: app}, nil
}
